# Repurika - レプリカ

Repurika is a Kubernetes Custom Resource works like ReplicaSet, an Kubernetes Native Resource. Repurika is developed using Kubebuilder, for Education purpose.

The improved version of Repurika is [Seika](https://github.com/bonavadeur/seika)

## 1. Demo

Install the CRD to Kubernetes System:

```bash
$ kubectl apply -f dist/install.yaml
```

Apply a Repurika (likes ReplicaSet)

```bash
$ cat dist/shuka-no-repurika.yaml
apiVersion: batch.bonavadeur.io/v1
kind: Repurika
metadata:
  labels:
    app.kubernetes.io/name: repurika
    app.kubernetes.io/instance: repurika-sample
    app.kubernetes.io/part-of: repurika
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: repurika
  name: shuka-no-repurika
spec:
  size: 2
  selector:
    matchLabels:
      app: shuka
  template:
    metadata:
      labels:
        app: shuka
    spec:
      containers:
      - name: shuka
        image: docker.io/bonavadeur/shuka:sleep

$ kubectl apply -f dist/shuka-no-repurika.yaml
repurika.batch.bonavadeur.io/shuka-no-repurika created

$ kubectl get pod | grep repurika
shuka-no-repurika-ytp55        1/1     Running   0              4m45s
shuka-no-repurika-zqu4h        1/1     Running   0              4m45s
```

Delete one `shuka-no-repurika-*`, you can see one more pod is created

## 2. Getting started from zero

Install Kubebuilder arcording site: [Kubebuilder Quick Start](https://book.kubebuilder.io/quick-start)

Scaffolding Out Our Project

```bash
# create a project directory, and then run the init command.
mkdir repurika
cd repurika
# scaffolding project
kubebuilder init --domain tesuto.bonavadeur.io --repo {your_repo}
```

## 3. Coding

Create atarashii API

```bash
kubebuilder create api --group batch --version v1 --kind Repurika
```

Define type in [api/v1/repurika_types.go](api/v1/repurika_types.go). Note that the field `RepurikaSpec.Template.Metadata` may be not embedded (don't worry about you do not understand this), means that you cannot apply the .yaml file with field `.spec.template.metadata`. Solution will be presented in following steps. Don't worry, I am here ^^.

Implement the Controller like file [internal/controller/repurika_controller.go](internal/controller/repurika_controller.go)

In this Project, I used my library for logging named `bonalib`, located in `internal/bonalib`, so I didn't use default library. To disable logging of default library, change this in [cmd/main.go](cmd/main.go)

```go
opts := zap.Options{
    Development: true,
    Level:       zapcore.Level(3),
}
```

## 4. Running

Whenever you change `*_types.go` file, you need to regenerate code by:

```bash
make generate
```

To generate manifest:

```bash
make manifests
```

Install CRD in Kubernetes system:

```bash
make install
```

In install step, you may be encounter error

```txt
The CustomResourceDefinition "repurikas.batch.bonavadeur.io" is invalid: metadata.annotations: Too long: must have at most 262144 bytes
```

Let's change Makefile, in PHONY manifests, add `crd:maxDescLen=0` option

```bash
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd:maxDescLen=0 webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

Uninstall CRD:

```bash
make uninstall
```

Run controller locally:

```bash
make run
```

Apply an instance of Custom Resource

```bash
kubectl apply -f config/samples/batch_v1_repurika.yaml
```

You can encounter an error in this step.

```text
Error from server (BadRequest): error when creating "config/samples/batch_v1_repurika.yaml": Repurika in version "v1" cannot be handled as a Repurika: strict decoding error: unknown field "spec.template.metadata.labels"
```

Let's enable embedded field in Makefile. Change PHONY manifests agains:

```bash
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd:maxDescLen=0,generateEmbeddedObjectMeta=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

This is the final version of Makefile :). Trust me!

Then, you can see colorful logging from terminal running `make run` command!

## 5. Packaging to ready to use anywhere!

Build image. Make sure that you have access to the Docker Repo

```bash
make docker-build docker-push docker.io/bonavadeur/repurika:latest
# change repo by yours, use full repo domain, include "docker.io", kudasai onegaishimasu.
```

You can deploy it now

```bash
make deploy IMG=docker.io/bonavadeur/repurika:latest
```

`Chotto matte kudasai!!!` Your product may be not work correctly! After you run `make deploy` command, your Kubernetes has a CR named repurika and you can list all repurika instance by command `kubectl get repurika`. But when apply a repurika instance, no Pod is created. Logging container `manager` in Pod `repurika-controll-manager` in namespace `repurika-system` you can see error:

```text
pkg/mod/k8s.io/client-go@v0.29.0/tools/cache/reflector.go:229: failed to list *v1.Pod: pods is forbidden: User "system:serviceaccount:repurika-system:repurika-controller-manager" cannot list resource "pods" in API group "" at the cluster scope
```

It means that your controller has no permission to list all Pods in Kubernetes system. Back to `internal/controller/repurika_controller.go` and add some markers:

```go
type RepurikaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.bonavadeur.io,resources=repurikas/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
```

Now, generate the installer file and bring it out over the World! Ikimashou!!!

```bash
make build-installer IMG=docker.io/bonavadeur/repurika:latest
```

Your installer file locate in `dist/install.yaml`

## 6. Épilogue

I hope that this is a small tutorial beside the tutorial on Kuberbuilder Book. `Repurika` is surely not a perfectly Kubebuilder Project. But I hope that it is useful for beginner (like me).

Any improvement can be sent to me, via `daodaihiep22ussr@gmail.com`, or create a new issues

## 7. Contributeur

Đào Hiệp - Bonavadeur - ボナちゃん  
The Future Internet Laboratory, E711 C7 Building, Hanoi University of Science and Technology, Vietnam.
未来のインターネット研究室, C7 の E ７１１、ハノイ百科大学、ベトナム。  

![](https://github.com/bonavadeur/bonavadeur/blob/master/images/github-wp.png?raw=true)
