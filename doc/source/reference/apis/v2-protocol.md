# V2 Inference Protocol

The V2 Inference Protocol is an industry-wide effort to provide an standardised
protocol to communicate with different inference servers (e.g. MLServer,
Triton, etc.) and orchestrating frameworks (e.g. Seldon Core, KServe, etc.).
The spec of the V2 Inference Protocol defines both the endpoints and payload
schemas for REST and gRPC interfaces.

As part of the V2 Protocol definition, you can find dedicated endpoints for:

- Health endpoints, to assess liveness and readiness of your model.
- Inference endpoints, to interact with your model.
- Metadata endpoints, to query your model metadata (e.g. expected inputs, expected
  outputs, etc.).
- Model repository endpoints, to load and unload models dynamically.


## REST

<div id="swagger-ui-rest"></div>
<script>
const HideHeaderPlugin = () => ({
   wrapComponents: {
      info: (Original, system) => (props) => null
   }
})

window.onload = function () {
   SwaggerUIBundle({
      url: "../../_static/openapi/v2/dataplane.yaml",
      dom_id: "#swagger-ui-rest",
      presets: [SwaggerUIBundle.presets.apis],
      plugins: [HideHeaderPlugin],
      docExpansion: "none",
      tryItOutEnabled: false
   });
};
</script>

## gRPC

.. mdinclude:: ../../../../proto/v2/README.md
