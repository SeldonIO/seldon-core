#TODO: want to this in deploy scripts rather than here
CLIENT_ID=sd-api
OIDC_USERNAME=admin@seldon.io
OIDC_PASSWORD=xxxxxx
OIDC_SCOPES='openid profile email groups'
DEPLOY_API_HOST=http://xx.xx.xx.xx/seldon-deploy/api/v1alpha1
OIDC_PROVIDER=http://xx.xx.xx.xx/auth/realms/deploy-realm

#TODO: something like make run_local but for just the user env get

kubectl create secret generic request-logger-auth -n seldon-logs \
  --from-literal=oidc_provider="${OIDC_PROVIDER}" \
  --from-literal=client_id="${CLIENT_ID}" \
  --from-literal=client_secret="${CLIENT_SECRET}" \
  --from-literal=oidc_scopes="${OIDC_SCOPES}" \
  --from-literal=oidc_username="${OIDC_USERNAME}" \
  --from-literal=oidc_password="${OIDC_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -

#TODO: put DEPLOY_API_HOST in env section of Deployment spec