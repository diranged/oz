// Package collector implements a Prometheus collector that reports the current
// inventory of Oz resources in the cluster.
//
// Counters and histograms live in the parent metrics package, because they are
// incremented from the code path that the event happens on. Point-in-time
// gauges ("how many access requests are live right now, for which template, for
// whom") cannot be maintained that way: a controller never reliably observes the
// disappearance of a resource, so an incrementally-maintained GaugeVec
// accumulates stale label sets forever.
//
// Instead this collector lists the resources from the manager's informer cache
// at scrape time. The cache is already populated by the controllers' Watches, so
// this costs no API calls, and the reported series cannot go stale.
//
// This package is separate from the parent metrics package to avoid an import
// cycle: the v1alpha1 API package imports metrics from its webhooks, and this
// collector imports the v1alpha1 API package for its types.
package collector

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/metrics"
)

// collectTimeout bounds how long a single scrape may spend listing resources.
// Reads are served from the informer cache so they should be near-instant; this
// exists only so that a wedged cache degrades the scrape rather than hanging it.
const collectTimeout = 10 * time.Second

var (
	requestsDesc = prometheus.NewDesc(
		"oz_access_requests",
		"Number of Oz access requests that currently exist, by kind, namespace, template, requesting user and readiness.",
		[]string{"kind", "namespace", "template", "user", "ready"},
		nil,
	)

	templatesDesc = prometheus.NewDesc(
		"oz_access_templates",
		"Number of Oz access templates that currently exist, by kind, namespace, name and readiness. Join against oz_access_requests to find templates nobody is using.",
		[]string{"kind", "namespace", "name", "ready"},
		nil,
	)
)

// Collector reports the current inventory of Oz access requests and templates.
type Collector struct {
	reader client.Reader
}

var _ prometheus.Collector = &Collector{}

// MustRegister builds a Collector backed by the supplied reader (normally
// mgr.GetClient(), so that reads hit the informer cache) and registers it
// against the controller-runtime metrics registry.
func MustRegister(reader client.Reader) {
	crmetrics.Registry.MustRegister(&Collector{reader: reader})
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- requestsDesc
	ch <- templatesDesc
}

// Collect implements prometheus.Collector.
//
// Any listing failure is reported as an invalid metric, which surfaces to
// Prometheus as a scrape error rather than as a silent zero. This is the
// expected behavior if a scrape somehow arrives before the informer cache has
// synced.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), collectTimeout)
	defer cancel()

	c.collectRequests(ctx, ch,
		&v1alpha1.PodAccessRequestList{},
		&v1alpha1.ExecAccessRequestList{},
	)
	c.collectTemplates(ctx, ch,
		&v1alpha1.PodAccessTemplateList{},
		&v1alpha1.ExecAccessTemplateList{},
	)
}

// requestKey is the label set for the oz_access_requests gauge.
type requestKey struct {
	kind      string
	namespace string
	template  string
	user      string
	ready     string
}

func (c *Collector) collectRequests(
	ctx context.Context,
	ch chan<- prometheus.Metric,
	lists ...client.ObjectList,
) {
	counts := map[requestKey]int{}

	for _, list := range lists {
		items, err := c.list(ctx, list)
		if err != nil {
			ch <- prometheus.NewInvalidMetric(requestsDesc, err)
			return
		}
		for _, item := range items {
			// Every request kind satisfies IRequestResource, which is where
			// the template name lives. Anything that does not is not something
			// we know how to report on.
			req, ok := item.(v1alpha1.IRequestResource)
			if !ok {
				continue
			}
			counts[requestKey{
				kind:      metrics.Kind(req),
				namespace: req.GetNamespace(),
				template:  req.GetTemplateName(),
				user:      metrics.UserLabel(req.GetAnnotations()[v1alpha1.AnnotationRequestedBy]),
				ready:     strconv.FormatBool(req.GetStatus().IsReady()),
			}]++
		}
	}

	for key, count := range counts {
		ch <- prometheus.MustNewConstMetric(
			requestsDesc,
			prometheus.GaugeValue,
			float64(count),
			key.kind, key.namespace, key.template, key.user, key.ready,
		)
	}
}

func (c *Collector) collectTemplates(
	ctx context.Context,
	ch chan<- prometheus.Metric,
	lists ...client.ObjectList,
) {
	// Templates are named and few, so unlike requests each one gets its own
	// series rather than being aggregated into a count. They are still buffered
	// until every list has succeeded, so that a failure partway through fails
	// the scrape outright instead of reporting a subset as though it were
	// everything.
	templates := []v1alpha1.ITemplateResource{}

	for _, list := range lists {
		items, err := c.list(ctx, list)
		if err != nil {
			ch <- prometheus.NewInvalidMetric(templatesDesc, err)
			return
		}
		for _, item := range items {
			tmpl, ok := item.(v1alpha1.ITemplateResource)
			if !ok {
				continue
			}
			templates = append(templates, tmpl)
		}
	}

	for _, tmpl := range templates {
		ch <- prometheus.MustNewConstMetric(
			templatesDesc,
			prometheus.GaugeValue,
			1,
			metrics.Kind(tmpl),
			tmpl.GetNamespace(),
			tmpl.GetName(),
			strconv.FormatBool(tmpl.GetStatus().IsReady()),
		)
	}
}

// list fetches a resource list and flattens it into the individual items.
// ExtractList takes the address of each item, so the returned objects satisfy
// the pointer-receiver Oz resource interfaces.
func (c *Collector) list(
	ctx context.Context,
	list client.ObjectList,
) ([]runtime.Object, error) {
	if err := c.reader.List(ctx, list); err != nil {
		return nil, err
	}
	return apimeta.ExtractList(list)
}
