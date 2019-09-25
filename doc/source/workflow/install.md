# Install Seldon-Core

**You will need a kubernetes cluster with version >=1.12**

## Webhook Certificates

For version >=0.5 Seldon Core requires certificates to be provisioned to the Webhook server. By default we assume the use of [cert-manager](https://github.com/jetstack/cert-manager). You can also provide your certificate in a secret directly if you don't want to use cert manager.

## Install Cert Manager

You can follow [the cert manager docmentation to install it](https://docs.cert-manager.io/en/latest/getting-started/install/kubernetes.html)

## Self-Signed Certificate

You will need to provide the certificate in a secret. The helm and kustomize installs in the following sections illustrate this route.

To install seldon-core on a Kubernetes cluster you have several choices:

We presently support [Helm](#seldon-core-helm-install) and [Kustomize](#seldon-core-kustomize-install).

## Seldon Core Helm Install

First [install Helm](https://docs.helm.sh). When helm is installed you can deploy the seldon controller to manage your Seldon Deployment graphs.

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system
```

**For the unreleased 0.5.0 version you would need to install 0.5.0-SNAPSHOT to test**:

```bash 
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --version 0.5.0-SNAPSHOT
```

Notes

 * You can use ```--namespace``` to install the seldon-core controller to a particular namespace but we recommend seldon-system.
 * For full configuration options see [here](../reference/helm.md)


### Install without cert-manager

To install with a provided certificate for the webhook you should create a values.yaml file containing the certificate details, e.g.:

```
webhook:
  ca:
    crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURQekNDQWllZ0F3SUJBZ0lVVDREc2ZoM3d2dERLVmNTbnM3aHE4RjJIUkRVd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0x6RXRNQ3NHQTFVRUF3d2tRV1J0YVhOemFXOXVJRU52Ym5SeWIyeHNaWElnVjJWaWFHOXZheUJFWlcxdgpJRU5CTUI0WERURTVNRGd6TURFMk5UVXhObG9YRFRFNU1Ea3lPVEUyTlRVeE5sb3dMekV0TUNzR0ExVUVBd3drClFXUnRhWE56YVc5dUlFTnZiblJ5YjJ4c1pYSWdWMlZpYUc5dmF5QkVaVzF2SUVOQk1JSUJJakFOQmdrcWhraUcKOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQW1Yck9HQVVVenlubGtUV2NMZXFvK3lmYW9Vb0l3OUFscHpTQgo4THVGUUR6cUZQMXJacUFCaVl6YkJMTjBLWGVHdGRBT2ZIK09NeU03b0lSR3VsaWpOOFMrRXQrRUk5ZmlJeUUwCnhtNm1kQjFiVlVhblBiMEVaeDk1QmNXWWVDcC8veE5PNkhyMmUrQlJLcnRWN2t5bUlveXlRNGk5RUhhaFNIR28KamdIeGhkendIMDNFRWJSdlo1M0VVMC9HMi9ySGpLY2tWdksvVHBPQktjVXNwKzJVV1dESXZVeE82eXhQQWdTTwpSN1Jyc0RiWXE5WWFuVng1bHJYSVdrOG53SW8zekg5eU95UXJsM0RSbTVDU04wcFEwTjNzdDFrSjhnUTRJRklGCmxOK3VxalVSeUJ3K2ZxSmZvRExlVzZ1OXA4RmhXUFdldXRWTmUyWVl1SnR6YXo4bHNRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVY01uUHNzNjcweGZPZ1hnZE9NL2UzV0lKZUFJd0h3WURWUjBqQkJnd0ZvQVVjTW5Qc3M2NwoweGZPZ1hnZE9NL2UzV0lKZUFJd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFRRUFSc2M0U3l3T3NxVWFqUjkwZS9hbVdwL0xoTVEwc2pieURhUzV4QU53bXQzL2lyUWtZUWRIenhsWFhFdHEKSWlaTzVSMnc5eTA4ZXcvNlgzWkhrSzBvWWhjMVc2WlhwOWVVdlI4MFc3TE9LV3BkTWtMNHQxMUhxM2VxZW80dQppd1lxaFUxdTRZUU8vdWVJbEdmbW9YZFlRQ29jbDgxdU03eXRza2hZbG85R1A5RUZCZ3ZLRExqOU5iZXd4TmFQCk1UUjN3MWdsQ0lROUtISGsxaUZnVGd1NkhLUXR6d3JHMFFkKzNMWnhqa1c4QXQ2RnRBZHRPbUtQUTBnaDZQd3cKbThpL2JLRFNxeHNoR1p0QzR3S2RhSXlERWpqMitoYjd5dlFkaXNycE13QVVuVGpJV2VySGZDdFk0YUdWTjlRRwpEa3o2SGdNbUxJZU0yc1E4Zkc3UGtOMWo1dz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  tls:
    crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM2VENDQWRFQ0ZEZXEyT2d0Z0JVRGdqYTEvM29tRmJ5TEs0KzdNQTBHQ1NxR1NJYjNEUUVCQ3dVQU1DOHgKTFRBckJnTlZCQU1NSkVGa2JXbHpjMmx2YmlCRGIyNTBjbTlzYkdWeUlGZGxZbWh2YjJzZ1JHVnRieUJEUVRBZQpGdzB4T1RBNE16QXhOalUxTVRaYUZ3MHhPVEE1TWpreE5qVTFNVFphTURNeE1UQXZCZ05WQkFNTUtITmxiR1J2CmJpMTNaV0pvYjI5ckxYTmxjblpwWTJVdWMyVnNaRzl1TFhONWMzUmxiUzV6ZG1Nd2dnRWlNQTBHQ1NxR1NJYjMKRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFEQTVialRDcmg3eUp0d0VsdHZmcCtzQXNEVW4zU0JkYlhkZFloSApOY1cwRGhqSTdSRXpQYndFVUNRTXJva05nVE1ua1RQZmhMTnZNa0p3eTNFVmJUQi93Tmtwdk94VEFPUnM0N0l1CkFLN0sxZmZzYXNhTnBvYS96czVTZzlBTENwRDVOVkFycWVCaW93cm1rRjFVRGY1Qm1rR1hueDhvNjhsQVNZNDEKbjRQWnFRSGhOZGIrSkQxRmhRS1dESk9rQXc2eHcreW5nMXpQbVF0MzRwakc4b2xJMDZqVkxWa0FKbW1saW50Uwo4ZjExa3B6UklOa0N1eWxUSDJjcVFoTkxDTHpNMk9NMXNmc1JxclQyY085WGJyM2ovL1pOUW1sMHA0SXVFV3Q2CjdsbnVFaU5GTTZPbyt2WmxWNFUxMUlSciswN3JrV0c3OURkSmE3NUhuR1IxdllBSEFnTUJBQUV3RFFZSktvWkkKaHZjTkFRRUxCUUFEZ2dFQkFBVSt5YzBtVmprREgrMloydzZEUWk5RXVHQVRjdDFPRjBHbGYzWTZBakJVYy9keAplZmNROUdaMUZ2ME9FVjM3RUx4Y09UQ25KcWM2TExMakI4V2YrdUFrL1Q2Rzh3MHpGN1FXdmRHVHU2UDdXRUd5ClNBWlkvYWx3RHEwUWpDUE1wWWNHNlJwdkRjREZBWktNRXJ5MGN1K1UwT3E4SmhNeVgyT1BJS3FyOUYvelp2RnQKWGRPdGU4MUR2VFNEbDFWRVhYRWFDcERHRDVPb1FDSGVqRzgzYTQyQVE1eHNVbzlCdWMwc2xsdWkvRE9jcWh5ZgprcGdEdEw2aVJsUDBXUEZ1TkRGTVhmSGFKL0cwSFp2eExGR3ZJNkZKV05SOVRRRi80Y2dsVitZUWNCcTdCaWhYCmJzRi96eDhGTFhGMVFnY2EwMjFwVjU2RFZnaXdsT1FBNXNvY0NIWT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBd09XNDB3cTRlOGliY0JKYmIzNmZyQUxBMUo5MGdYVzEzWFdJUnpYRnRBNFl5TzBSCk16MjhCRkFrREs2SkRZRXpKNUV6MzRTemJ6SkNjTXR4Rlcwd2Y4RFpLYnpzVXdEa2JPT3lMZ0N1eXRYMzdHckcKamFhR3Y4N09Vb1BRQ3dxUStUVlFLNm5nWXFNSzVwQmRWQTMrUVpwQmw1OGZLT3ZKUUVtT05aK0QyYWtCNFRYVwovaVE5UllVQ2xneVRwQU1Pc2NQc3A0TmN6NWtMZCtLWXh2S0pTTk9vMVMxWkFDWnBwWXA3VXZIOWRaS2MwU0RaCkFyc3BVeDluS2tJVFN3aTh6TmpqTmJIN0VhcTA5bkR2VjI2OTQvLzJUVUpwZEtlQ0xoRnJldTVaN2hJalJUT2oKcVByMlpWZUZOZFNFYS90TzY1Rmh1L1EzU1d1K1I1eGtkYjJBQndJREFRQUJBb0lCQUNndDhaQ3NFLzljcXR2dQpSdk56YWFqM3JkamNHZlY1WWxkdHl1UWlWRHNNRUtlUmtkcWRpbE5QcWlLbUhGQWUwRnYyaDlxUUZwd2IyUEVMCnYxTmFPaGJ5UVluTEcyS3l0ZUhrajlHN1BLMXRZa1h4ZThnM25xdkhWUHlsRGltdW1zSCtFK1AwYjVPOEtHSWMKUWdSbkljWGliclU1Wk5FdVErNUxJLzhSYWZKbFAvZEJVNzBQTHIzd3lvM0xwNjRHSmF3Tk5NV3pKR0ErZ0J3MgpQS3pwU0xDMGM0SEd5TERFbnI2bjRtZm9HT3ZlT0xLNWJnWDVCSWhBS3ZUeFBNKzBwK3lIL0lTdGlNUWhYVHNvCnBZdzJDT3hLckw5Y0tRRVFzVUNDaEFwazNXMGN2YXNqaGExMlY0b3h4a0NiRmN3SVphaUF4V1I1ZEQreXVDbG4KVkd2dmRDa0NnWUVBN3hGM0RxZEtCdTFnQU0vSWZiZDRPeG1JVUVkVWZpd3AvQnJmNThxVXJ1bE9XdlUybXVkdwpOU0xoMmp0SVRhdUNndmhpdUk1bU15RjQ4bHAwNGZwQjMvZXc5Y1dGSldDS1UvKzc4Q3pzSk1tN0FJNGlHMDFQCmVicDNkc3VFeGlZdHE2eXFFS05MZDhzRlZ3OWRxSjBFUmhkZ2Z5V3dtZVAvWXhLVVJseWt5RDBDZ1lFQXpvOGkKakx2OWFuQTdnR0tXdU51Y2hEbzA0Ym9SblA2MVB5SVUyTGlYNGdaUEgwbWgrQ1ZtM3BQNWtTTXErWGluRGJ1cwpQYTdnN1JmMnhiclFKRG9Ub0Y1cGtCS0FmQVUyb2MxdStrL0gxUmFHdWdrbGxqeWRuc1lCNERDTWV2aUVNVHh2CjFHQ2VsRUN4alBwalQydDA4d1ZRK1VyaERaek5WT3dxa2d0TDZaTUNnWUJsaytBb1k4QTZiVVdyVXAzM2ZLc2oKUVZmLzlDN2NaVnQ1ZU5uR0hQZEwwbW11a0I0aGQxRGY0dkJmejJ5TFErSnlUNk55azE2dFB2MnF5L0I1eStHTgpqaXFzWXI2T0FSVUZWOVc4MlBtRk1BbTYxS2w5UEQ0V2xMb0p5Yk9pbGJvMkJXbEZKSHorYTA3YmpQWFluTTZpCkVYQzQxWVRSL21RVzdsLzkvWU11YVFLQmdFSzFMUlpBTy9ZYzZzcHFqSHlFeUFaWCtlNFFObEg2WERSWVlGMGgKT0VQUmY4bjk4S1lBQmpuSmxpYU9NZm5CUWtvSUd2Y012QzAxdVFkZ2JvblVpN1FWNlllU3doWExaVHBaNndaQgpyNnFjak1RVjRpS2p6cytRNk5nck5hTWRFU3dKZGFBajEvTE85Y2d1c05YY1FUZWV0dWpiaXRUbmw5UmVOTjFYCmNwdXJBb0dBQ1gvcWthSEZoYXltRkNlZmtmRjdwbmJsN0tQNVRHa2g0WmVUSGk3OEM5WE5yS21GQkFYSXV6dEoKRXJTdjREdU5oa2lGVW5qc21SVmhKaU5WQ2p0UDQzMVB1cXlnWXJneDNkOTZlYm9oRVN4MTBXMlBtdHg5WTZQagp3V29iOHZBVzVERFRqdFhRNkRhUjNyUHVmNmJaQ2dmTzUxcTZPWTdPTTYvUkFLci9ndTg9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==

```

You can then install using these values with:

```
helm install -f seldon-core-operator/values-self-signed-cert.yaml seldon-core-operator --name seldon-core  --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system --version 0.5.0-SNAPSHOT --set webhook.certManager.enabled=false --set webhook.secretProvided=true
```

## Ingress Support

For particular ingresses we support you can inform the controller it should activate processing for them.

 * Ambassador
   * add `--set ambassador.enabled=true` : The controller will add annotations to services it creates so Ambassador can pick them up and wire an endpoint for your deployments.
 * Istio Gateway
   * add `--set istio.enabled=true` : The controller will create virtual services and destination rules to wire up endpoints in your istio ingress gateway.

## Install an Ingress Gateway

We presently support two API Ingress Gateways

 * [Ambassador](https://www.getambassador.io/)
 * [Istio Ingress](https://istio.io/)

### Install Ambassador

We suggest you install [the official helm chart](https://github.com/helm/charts/tree/master/stable/ambassador). At present we recommend 0.40.2 version due to issues with grpc in the latest.

```
helm install stable/ambassador --name ambassador --set crds.keep=false
```

### Install Istio Ingress Gateway

If you are using istio then the controller will create virtual services for an istio gateway. By default it will assume the gateway `seldon-gateway` as the name of the gateway. To change the default gateway add `--set istio.gateway=XYZ` when installing the seldon-core-operator.


## Seldon Core Kustomize Install 

The [Kustomize](https://github.com/kubernetes-sigs/kustomize) installation can be found in the `/operator/config` folder of the repo. You should copy this template to your own kustomize location for editing.

To use the template directly there is a Makefile which has a set of useful commands:


Install cert-manager

```
make install-cert-manager
```

Install Seldon using cert-manager to provide certificates.

```
make deploy
```

Install Seldon with provided certificates in `config/cert/`

```
make deploy-cert
```


## Other Options

### Install with Kubeflow

  * [Install Seldon as part of Kubeflow.](https://www.kubeflow.org/docs/guides/components/seldon/#seldon-serving)

### GCP MarketPlace

If you have a Google Cloud Platform account you can install via the [GCP Marketplace](https://console.cloud.google.com/marketplace/details/seldon-portal/seldon-core).

## Upgrading from Previous Versions

See our [upgrading notes](../reference/upgrading.md)

