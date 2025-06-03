# sys11iam Provider

The sys11iam provider is used to interact with the SysEleven IAM. The provider needs to be configured with the proper credentials (a service account) before it can be used.

## Example Usage

```hcl
# Define required providers
terraform {
  required_providers {
    sys11iam = {
      source = "syseleven/sys11iam"
    }
  }
}

Using the `sys11iam` provider above:

```hcl
# Configure the sys11iam Provider for service account user authentication
provider "ncs" {
  serviceaccount_secret = "s11_orgsa_asdziuch-967s-aduc-123f-00asdasd8asd_9xjakshdkjOJPvk-36FJqasdmalkwaksnkajc"
  iam_url = "https://iam.apis.syseleven.de"
}
```

## Configuration Reference

The following arguments are supported for the provider "sys11iam":

* **`iam_url`** - The url to the IAM service for creating organization, project, organization membership and project membership resources.
  If omitted, the `SYS11IAM_IAM_URL` environment variable is used.
* **`serviceaccount_secret`** - The secret of an service account to authenticate with. If omitted, the `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.
  

