behavior "remove_labels_on_reply" "remove_stale" {
    labels = ["waiting-response", "stale"]
    only_non_maintainers = true
}

poll "label_issue_migrater" "provider_migrater" {
    schedule = "0 40 * * * *"
    new_owner = env.PROVIDERS_OWNER
    repo_prefix = "terraform-provider-"
    label_prefix = "provider/"
    issue_header = <<-EOF
    _This issue was originally opened by @${var.user} as ${var.repository}#${var.issue_number}. The original body of the issue is below._
    
    <hr>
    
    EOF
    migrated_comment = "This issue has been automatically migrated to ${var.repository}#${var.issue_number} because it looks like an issue with that provider. If you believe this is _not_ an issue with the provider, please reply to ${var.repository}#${var.issue_number}."
}

poll "closed_issue_locker" "locker" {
  schedule             = "0 50 1 * * *"
  closed_for           = "720h" # 30 days
  max_issues           = 500
  sleep_between_issues = "5s"

  message = <<-EOF
    I'm going to lock this issue because it has been closed for _30 days_ ⏳. This helps our maintainers find and focus on the active issues.

    If you have found a problem that seems similar to this, please open a new issue and complete the issue template so we can capture all the details necessary to investigate further.
  EOF
}
