# Open Inference Protocol

The Open Inference Protocol is an industry-wide effort to provide a standardized
protocol to communicate with different inference servers (e.g. MLServer,
Triton, etc.) and orchestrating frameworks (e.g. Seldon Core, KServe, etc.).
The spec of the Open Inference Protocol defines both the endpoints and payload
schemas for REST and gRPC interfaces.

As part of the Open Inference Protocol definition, you can find dedicated endpoints for:

- Model controls: Call model inference, interact with your model, and load and unload models dynamically
- Health: Assess liveness and readiness of your model.
- Metadata: Query your model metadata (e.g. expected inputs, expected
  outputs, etc.).


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
