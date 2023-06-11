package controller

var createJobQuery = `
mutation CreateJob($input: CreateJobInput!) {
	createJob(input: $input) {
	  ... on CreateJobSuccess {
		job {
		  id
		  __typename
		}
		__typename
	  }
	  ... on InputErrors {
		errors {
		  path
		  message
		  code
		  __typename
		}
		__typename
	  }
	  __typename
	}
  }
`
