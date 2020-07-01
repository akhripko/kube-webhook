# Description
The main goal is providing an opportunity to separate access in the kubernetes cluster by service owners.

k8s-acl-sv is webhook admission controller that adds special annotation 'owners' and could validate access to that resource only for owners of it.  

# Kubernetes admission controllers
Kubernetes admission controllers are plugins that govern and enforce how the cluster is used. They can be thought of as a gatekeeper that intercept (authenticated) API requests and may change the request object or deny the request altogether. 

Admission webhooks are HTTP callbacks that receive admission requests and do something with them. You can define two types of admission webhooks, validating admission webhook and mutating admission webhook. 
(see .helm/templates/webhook_*.yaml) 

There are next operations of webhook requests:
- CONNECT
- CREATE
- UPDATE
- DELETE


`CONNECT and DELETE should be used as validation webhook (see kind: ValidatingWebhookConfiguration)`<br>
`CREATE and UPDATE should be used as mutation webhook (see kind: MutatingWebhookConfiguration)`

# Internal logic

There are 3 webhook endpoints:

- `/add-owners`: add current user to the list of owners 
- `/mutate`: check access of CREATE & UPDATE operations & modify owners list according to the business logic
- `/validate`: check access of CONNECT & DELETE operations

There are main configs of mutation (.helm/templates/webhook_mutate.yaml) & validation (.helm/templates/webhook_validate.yaml)<br> 
In order to change webhook logic the path (webhooks>clientConfig>service>path) should have an appropriate value.

Webhook `rules` is a part of webhook config where the scope of resources could be defined that will be managed by webhook:<br>
 
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["*"]
        apiVersions: ["*"]
        resources: ["deployments", "services"]
        scope: "Namespaced"

Note: 
    
    The scope field specifies if only cluster-scoped resources ("Cluster") or namespace-scoped resources 
    ("Namespaced") will match this rule. "*" means that there are no scope restrictions.

# Generic make commands 

build & deploy to an appropriate cluster
- `local-dev` (in case of custom dev branch)
- `local` 
- `stage-trading` 
- `production-trading` 
- `stage-betting` 
- `production-betting`

see make-vars.mk, .helm/values-*.yaml

# References
- https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/
- https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
