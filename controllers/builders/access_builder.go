package builders

type AccessBuilder struct {
	*BaseBuilder
}

// TODO: GeneratePodName needs to figure out the PodName after it has created the target pod in the first place? Or
// it could just generate a static name with a clean function and return that.
func (t *AccessBuilder) GeneratePodName() (string, error) {
	return "junk", nil
}
