structure helm chart for 1 server manage many chart
``gitops/
 ├── apps/
 │    ├── api/
 │    │    ├── Chart.yaml
 │    │    ├── templates/
 │    │    │    ├── deployment.yaml
 │    │    │    ├── service.yaml
 │    │    │    ├── ingress.yaml
 │    │    │    └── configmap.yaml
 │    │    └── values.yaml
 │    │
 │    └── cms/
 │         ├── Chart.yaml
 │         ├── templates/
 │         │    ├── deployment.yaml
 │         │    ├── service.yaml
 │         │    ├── ingress.yaml
 │         │    └── configmap.yaml
 │         └── values.yaml
 │
 └── envs/
      ├── site1/
      │    ├── api/values.yaml
      │    └── cms/values.yaml
      ├── site2/
      │    ├── api/values.yaml
      │    └── cms/values.yaml
      └── clientA/
           ├── api/values.yaml
           └── cms/values.yaml
``
