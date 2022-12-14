package cmd

// Used after an Access Request has been successfully created.
//
// Eg:
//
//	  if req.GetStatus().IsReady() {
//		cmd.Printf(logSuccess(successMsg), req.GetStatus().GetAccessMessage())
//		break
//	  }
var successMsg = logSuccess(`
Success, your access request is ready! Here are your access instructions:
---
%s
---
`)
