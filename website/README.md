# Terraform Documentation

This directory contains the portions of [the Terraform website][terraform.io] that pertain to the Terraform Plugin SDK.

The files in this directory are intended to be used in conjunction with
[the `terraform-website` repository](https://github.com/hashicorp/terraform-website), which brings all of the
different documentation sources together and contains the scripts for testing and building the site as
a whole.

## Updating Sidebar Navigation

You must update the sidebar navigation for the `terraform-plugin-sdk` documentation any time that you add or delete a documentation page. The website builds the sidebar navigation menu from the [nav-data] JSON file. For more details about how to update this file, refer to https://github.com/hashicorp/terraform-website#editing-navigation-sidebars.

## Adding Redirects

You must add a redirect when you move, rename, or delete documentation pages. Refer to https://github.com/hashicorp/terraform-website#redirects for details.

## Previewing Changes

You should preview your changes locally to ensure that the content is rendering properly before you create a pull request. The build includes content from this repository and the [`terraform-website`](https://github.com/hashicorp/terraform-website/) repository, allowing you to preview the entire Terraform documentation site.

To preview your content, complete the following steps:

**Set Up Local Environment**

1. [Install Docker](https://docs.docker.com/get-docker/).
1. Restart your terminal or command line session.

**Launch Site Locally**

1. Navigate into your local `terraform-plugin-sdk` top-level directory and run `make website`.
1. Open `http://localhost:3000` in your web browser. While the preview is running, you can edit pages and Next.js will automatically rebuild them.
1. When you're done with the preview, press `ctrl-C` in your terminal to stop the server.

### Validating Content

Content changes are automatically validated against a set of rules as part of the pull request process. If you want to run these checks locally to validate your content before committing your changes, you can run the following command:

```
npm run content-check
```

If the validation fails, actionable error messages will be displayed to help you address detected issues.

## Deployment

The website reads content from release tags to generate documentation for all versions of `terraform-plugin-sdk` documentation. Changes merged into `main` will be included in the documentation for the next product release.

You cannot edit documentation for past versions of `terraform-plugin-sdk` on the site. Documentation is an artifact of a product release. We push docs fixes forward for the next release, rather than retroactively fixing older versions.

[nav-data]: ../website/data/plugin-sdk-nav-data.json
[terraform.io]: https://www.terraform.io/
