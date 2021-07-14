
# Super Simple Mesh
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-81%25-brightgreen.svg?longCache=true&style=flat)</a>

A very simple service mesh providing only inter-services TLS.
Super Simple Mesh is useful for teams who just want to encrypt flows between their containers without bothering with the extra features and complexity a classic Service Mesh brings


SSM needs Cert-Manager for delivering certificates. It will request certificates for your workloads based on annotations that you provide.

### Setup
```
git clone git@github.com:etiennejournet/Super-Simple-Mesh.git
kubectl apply -f Super-Simple-Mesh/deploy/manifest
```

### With Cert-Manager 
#### Basic setup
Setup Cert-Manager, using helm for example :

    helm install jetstack/cert-manager

Define a CA Cluster Issuer according to [this documentation.](https://cert-manager.io/docs/configuration/ca/)
Note that SSM will use a Cluster Issuer called "caIssuer" by default, refer to the annotation list for another behavior
 
#### Annotations 
| Annotation Name | Description | Required| Default |
|--|--|--|--|
| cert-manager.ssm.io/service-name | Name of the Service for which your pods are going to be certified for | Yes | None
| cert-manager.ssm.io/cluster-issuer | Name of the Cluster Issuer that SSM should request a Certificate to | No | caIssuer
| cert-manager.ssm.io/cert-duration | Duration of the requested certificate | No | 24h 

