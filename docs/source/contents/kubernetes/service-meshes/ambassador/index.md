# Ambassador

[Ambassador](https://www.getambassador.io/) provides service mesh and ingress products. Our examples here are based on the Emissary ingress.

We will run through some examples as shown in the notebook `service-meshes/ambassador/ambassador.ipynb`

## Single Model

 * Seldon Iris classifier model
 * Default Ambassador Host and Listener
 * Ambassador Mappings for REST and gRPC endpoints

```{literalinclude} ../../../../../../service-meshes/ambassador/static/single-model.yaml 
:language: yaml
```

## Traffic Split

```{warning}
Traffic splitting does not presently work due to this [issue](https://github.com/emissary-ingress/emissary/issues/4062). We recommend you use a Seldon Experiment instead.
```

Seldon provides an Experiment resource for service mesh agnostic traffic splitting but if you wish to control this via Ambassador and example is shown below to split traffic between two models.

```{literalinclude} ../../../../../../service-meshes/ambassador/static/traffic-split.yaml 
:language: yaml
```



```{include} ../../../../../../service-meshes/ambassador/README.md
```

