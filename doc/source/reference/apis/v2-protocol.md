# V2 Inference Protocol

The V2 Inference Protocol is an industry-wide effort to provide an standardised
protocol to communicate with different inference servers (e.g. MLServer,
Triton, etc.) and orchestrating frameworks (e.g. Seldon Core, KServe, etc.).
The spec for the V2 Inference Protocol defines both the endpoints and payload
schemas for REST and gRPC interfaces.

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
      url: "../../_static/dataplane.yaml",
      dom_id: "#swagger-ui-rest",
      presets: [SwaggerUIBundle.presets.apis],
      plugins: [HideHeaderPlugin],
      docExpansion: "none",
      tryItOutEnabled: false
   });
};
</script>

## gRPC

bli bla blu

