
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

### Configuration 

Super Simple Mesh is configured using environnement variables 
| Environnement Variable | Description | Required| Default |
|--|--|--|--|
| CERTMANAGER_ISSUER | Issuer that should be used for delivering certificates. For now SSM only supports ClusterIssuers | No | 777
| WEBHOOK_NAME | Name of the SSM service (only useful if you changed it in the YAML manifest) | No | ssm
| WEBHOOK_PORT | Port of the SSM service (only useful if you changed it in the YAML manifest) | No | 8080
| ENVOY_UID | User ID of the Envoy Proxy User, change it to a unique value if the default is already used | No | 777 

### With Cert-Manager 
#### Basic setup
Setup Cert-Manager, using helm for example :

    helm install jetstack/cert-manager

Define a CA Cluster Issuer according to [this documentation.](https://cert-manager.io/docs/configuration/ca/)
Note that SSM will use a Cluster Issuer called "caIssuer" by default, refer to the annotation list for another behavior
 
#### Annotations 
| Annotation Name | Description | Required| Default |
|--|--|--|--|
| cert-manager.ssm.io/service-name | Name of the Service for which the pods should be certified for | Yes | None
| cert-manager.ssm.io/cert-duration | Duration of the requested certificate | No | 24h 

