# Tarian

> ###### We want to maintain this as an open-source project to fight against the attacks on our favorite Kubernetes ecosystem. By continuous contribution, we can fight threats together as a community.


##### Protect your Applications running on Kubernetes from malicious attacks by pre-registering your source code signatures, runtime processes monitoring, runtime source code monitoring, change detection, alerting, pre-configured & instant respond actions based on detections and also sharing detections with community. Save your K8s environment from Ransomware! 

#

[![Build status](https://img.shields.io/github/workflow/status/kube-tarian/tarian/CI?style=flat)](https://github.com/kube-tarian/tarian/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kube-tarian/tarian)](https://goreportcard.com/report/github.com/kube-tarian/tarian)

#

How does Tarian work?
> Tarian runs as a sidecar container in your main application's pod detecting unknown processes and changes applied to your files' signatures, etc. Tarian will be a part of your Application's pod from dev to prod environment, hence you can register to your Tarian DB what is supposed to be happening & running in your container + file signatures to be watched + what can be notified + action to take (self destroy the pod) based on changes detected. Shift-left your detection mechanism!

What if an unknown change happens inside the container which is not in Tarian's registration DB, how does Tarian react to it? 
> If an unknown change happens, Tarian can simply notify observed analytics to your Security Team. Then your Security Engineers can register that change in Tarian DB whether it's considered a threat or not. Also, based on their analysis they can configure what action to take when that change happens again. That action will be sent as a command to the sidecar Tarian app to be performed.

How does the contribution of community helps to fight against the threats via Tarian?
> Any new detection analyzed & marked as a threat by your Security Experts, if they choose, can be shared to the open-source Tarian community DB with all the logs, strings to look for, observation, transparency, actions to configure, ... Basically anything the Experts want to warn about & share with the community. You can use that information as a Tarian user and configure actions in the Tarian app which is used in your environment. This is basically a mechanism to share info about threats & what to do with them. This helps everyone using Tarian to take actions together in their respective K8s environments by sharing their knowledge & experience.

What kind of action(s) would Tarian take based on known threat(s)?
> Tarian would simply self destroy the pod it's running on. If the malware/virus spreads to the rest of the environment, well you know what happens. So, Tarian is basically designed to help reduce the risk as much as possible by destroying pods. Provisioning a new pod will be taken care of by K8s since that's how K8s works. Tarian will only do destruction of the pods only if you tell Tarian to do so. If you don't want any actions to happen, you don't have to configure or trigger any; you can simply tell Tarian to just notify you. Tarian basically does what you want to be done to reduce the risk.

#

Why another new security tool when there are many tools available already, like Falco, Kube-Hunter, Kube-Bench, Calico Enterprise Security, and many more security tools (open-source & commercial) that can detect & prevent threats at network, infra & application level? Why Tarian?
> Like I mentioned above, the main reason Tarian was born is to fight against threats in Kubernetes together as a community. Another reason was, what if there is still some sophisticated attack which is capable of penetrating every layer of your security, able to reach your runtime app, your storage volumes and capable of spreading to damage or lock your infra & data?! What do you want to do about such attacks, especially which turns into ransomware. Tarian is designed to reduce such risks, by taking action(s). We know that Tarian is not the ultimate solution, but we are confident that it can help reduce risks especially when knowledge is shared continuosly by the community. From a technical perspective, Tarian can help reduce the risk by destroying the infected resources.

#

#### Architecture diagram

![Arch. Diagram](https://github.com/kube-tarian/tarian/blob/5eeed9a0bd5875e6cee423d2d12161a3f7d2d84c/Kube-Tarian.svg)

#

## Prerequisites

A kubernetes cluster that supports running [Falco](https://falco.org/docs/getting-started/)

## Install

Tarian integrates with Falco by subscribing Falco Alerts via [gRPC API](https://falco.org/docs/grpc/). Falco support running gRPC API with mandatory mutual TLS (mTLS). So, firstly we need to prepare the certificates.

### Prepare Namespaces

```bash
kubectl create namespace tarian-system
kubectl create namespace falco
```

### Prepare Certificate for mTLS

#### With Cert Manager

You can setup certificates manually and save those certs to secrets accessible from Falco and Tarian pods. For convenient, you can use Cert Manager to manage the certs.

1. Install Cert Manager by following this guide https://cert-manager.io/docs/installation/
2. Wait for cert manager pods to be ready

```bash
kubectl wait --for=condition=ready pods --all -n cert-manager --timeout=3m
```

3. Setup certs

##### A. If you don't have an existing cluster issuer, you can create one using a self-signed issuer

Save this to `tarian-falco-certs.yaml`, then run `kubectl apply -f tarian-falco-certs.yaml`.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: root-ca
  namespace: cert-manager
spec:
  isCA: true
  commonName: root-ca
  secretName: root-secret
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: ca-issuer
spec:
  ca:
    secretName: root-secret
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: falco-grpc-server
  namespace: falco
spec:
  isCA: false
  commonName: falco-grpc
  dnsNames:
  - falco-grpc.falco.svc
  - falco-grpc
  secretName: falco-grpc-server-cert
  usages:
  - server auth
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: ca-issuer
    kind: ClusterIssuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: falco-integration-cert
  namespace: tarian-system
spec:
  isCA: false
  commonName: tarian-falco-integration
  dnsNames:
  - tarian-falco-integration
  usages:
  - client auth
  secretName: tarian-falco-integration
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: ca-issuer
    kind: ClusterIssuer
    group: cert-manager.io
```

##### B. If you have an existing cluster issuer

Save this to `tarian-falco-certs.yaml`, then run `kubectl apply -f tarian-falco-certs.yaml`.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: falco-grpc-server
  namespace: falco
spec:
  isCA: false
  commonName: falco-grpc
  dnsNames:
  - falco-grpc.falco.svc
  - falco-grpc
  secretName: falco-grpc-server-cert
  usages:
  - server auth
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: your-issuer # change this to yours
    kind: ClusterIssuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: falco-integration-cert
  namespace: tarian-system
spec:
  isCA: false
  commonName: tarian-falco-integration
  dnsNames:
  - tarian-falco-integration
  usages:
  - client auth
  secretName: tarian-falco-integration
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: your-issuer # change this to yours
    kind: ClusterIssuer
    group: cert-manager.io
```

#### Setup certificates manually

If you have other ways to setup the certificates, that would work too. You can create kubernetes secrets containing those certificates.
The following steps expect that the secrets are named:

- `tarian-falco-integration` in namespace `tarian-system`
- `falco-grpc-server-cert` in namespace `falco`

For mTLS to work, those certificates need to be signed by the same CA.


### Install Falco with custom rules from Tarian

Save this to `falco-values.yaml`

```yaml
extraVolumes:
- name: grpc-cert
  secret:
    secretName: falco-grpc-server-cert
extraVolumeMounts:
- name: grpc-cert
  mountPath: /etc/falco/grpc-cert
falco:
  grpc:
    enabled: true
    unixSocketPath: ""
    threadiness: 1
    listenPort: 5060
    privateKey: /etc/falco/grpc-cert/tls.key
    certChain: /etc/falco/grpc-cert/tls.crt
    rootCerts: /etc/falco/grpc-cert/ca.crt
  grpcOutput:
    enabled: true
```

Then install Falco using Helm:

```bash
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update

helm upgrade -i falco falcosecurity/falco -n falco -f falco-values.yaml \
  --set-file customRules."tarian_rules\.yaml"=https://raw.githubusercontent.com/kube-tarian/tarian/main/dev/falco/tarian_rules.yaml
```

### Setup a Postgresql Database

You can use a DB as a service from your Cloud Services or you can also run by yourself in the cluster. For example to install the DB in the cluster, run:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install tarian-postgresql bitnami/postgresql -n tarian-system \
  --set postgresqlUsername=postgres \
  --set postgresqlPassword=tarian \
  --set postgresqlDatabase=tarian
```

### Install tarian

1. Install tarian using Helm

```bash
helm repo add tarian https://kube-tarian.github.io/tarian
helm repo update

helm upgrade -i tarian-server tarian/tarian-server --devel -n tarian-system
helm upgrade -i tarian-cluster-agent tarian/tarian-cluster-agent --devel -n tarian-system
```

2. Wait for all the pods to be ready

```bash
kubectl wait --for=condition=ready pod --all -n tarian-system
```

3. Run database migration to create the required tables

```bash
kubectl exec -ti deploy/tarian-server -n tarian-system -- ./tarian-server db migrate
```

4. Verify

After the above step, you should see falco alert in tarianctl get events (See the following Usage sections).


## Configuration

See helm chart values for
- [tarian-server](https://github.com/kube-tarian/tarian/blob/main/charts/tarian-server/values.yaml)
- [tarian-cluster-agent](https://github.com/kube-tarian/tarian/blob/main/charts/tarian-cluster-agent/values.yaml)


## Cloud / Vendor specific configuration

### Private GKE cluster

Private GKE cluster by default creates firewall rules to restrict master to nodes communication only on ports `443` and `10250`.
To inject tarian-pod-agent container, tarian uses a mutating admission webhook. The webhook server runs on port `9443`. So, we need
to create a new firewall rule to allow ingress from master IP address range to nodes on tcp port **9443**.

For more details, see GKE docs on this topic: [https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters#add_firewall_rules](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters#add_firewall_rules).


## Usage

### Use tarianctl to control tarian-server

1. Download from Github [release page](https://github.com/kube-tarian/tarian/releases)
2. Extract the file and copy tarianctl to your PATH directory
3. Expose tarian-server to your machine, through Ingress or port-forward. For this example, we'll use port-forward:

```bash
kubectl port-forward svc/tarian-server -n tarian-system 41051:80
```

4. Configure server address with env var

```
export TARIAN_SERVER_ADDRESS=localhost:41051
```

### To see violation events

```bash
tarianctl get events
```

### Add a process constraint

```bash
tarianctl add constraint --name nginx --namespace default \
  --match-labels run=nginx \
  --allowed-processes=pause,tarian-pod-agent,nginx 
```

```bash
tarianctl get constraints
```

### Add a file constraint

```bash
tarianctl add constraint --name nginx-files --namespace default \
  --match-labels run=nginx \
  --allowed-file-sha256sums=/usr/share/nginx/html/index.html=38ffd4972ae513a0c79a8be4573403edcd709f0f572105362b08ff50cf6de521
```

```bash
tarianctl get constraints
```

### Run tarian agent in a pod

Then after the constraints are created, we inject tarian-pod-agent to the pod by adding an annotation:

```yaml
metadata:
  annotations:
    pod-agent.k8s.tarian.dev/threat-scan: "true"
```

Pod with this annotation will have an additional container injected (tarian-pod-agent). The tarian-pod-agent container will 
continuously verify the runtime environment based on the registered constraints. Any violation would be reported, which would be
accessible with `tarianctl get events`.


### Demo: Try a pod that violates the constraints

```bash
kubectl apply -f https://raw.githubusercontent.com/kube-tarian/tarian/main/dev/config/monitored-pod/configmap.yaml
kubectl apply -f https://raw.githubusercontent.com/kube-tarian/tarian/main/dev/config/monitored-pod/pod.yaml

# wait for it to become ready
kubectl wait --for=condition=ready pod nginx

# simulate unknown process runs
kubectl exec -ti nginx -c nginx -- sleep 15

# you should see it reported in tarian
tarianctl get events
```

## Alert Manager Integration

Tarian comes with Prometheus Alert Manager by default. If you want to use another alert manager instance:

```bash
helm install tarian-server tarian/tarian-server --devel \
  --set server.alert.alertManagerAddress=http://alertmanager.monitoring.svc:9093 \
  --set alertManager.install=false \
  -n tarian-system
```

To disable it, you can set the alertManagerAddress value to empty.

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md)

## Automatic Constraint Registration

When tarian-pod-agent runs in registration mode, instead of reporting unknown processes and files as violations, it automatically registers them as a new constraint. This is convenient to save time from registering manually.

To enable constraint registration, the cluster-agent needs to be configured.

```bash
helm install tarian-cluster-agent tarian/tarian-cluster-agent --devel -n tarian-system \
  --set clusterAgent.enableAddConstraint=true
```

```yaml
metadata:
  annotations:
    # register both processes and file checksums
    pod-agent.k8s.tarian.dev/register: "processes,files"
    # ignore specific paths from automatic registration
    pod-agent.k8s.tarian.dev/register-file-ignore-paths: "/usr/share/nginx/**/*.txt"
```

Automatic constraint registration can also be done in a dev/staging cluster, so that there would be less changes in production.

## Other supported annotations

```yaml
metadata:
  annotations:
    # specify how often tarian-pod-agent should verify file checksum
    pod-agent.k8s.tarian.dev/file-validation-interval: "1m"
```

## Securing tarian-server with TLS

To secure tarian-server with TLS, create a secret containing the TLS certificate. You can create the secret manually,
or using [Cert Manager](https://cert-manager.io/). Once you have the secret, you can pass the name to the helm chart value:

```
helm upgrade -i tarian-server tarian/tarian-server --devel -n tarian-system \
  --set server.tlsSecretName=tarian-server-tls
```

## Contributing

See [docs/contributing.md](docs/contributing.md)

