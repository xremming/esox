output "function_urls" {
  value = { for env, _ in local.homepage : env => module.homepage[env].function_url }
}
