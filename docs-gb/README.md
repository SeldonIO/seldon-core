# About

Seldon Core 2 is a source-available, Kubernetes-native framework designed to deploy and manage machine learning (ML) systems at scale. Its data-centric approach and modular architecture enable organizations to handle everything from simple models to complex ML applications, ensuring flexibility, observability, and cost efficiency across diverse environments, including on-premise, hybrid, and multi-cloud setups.

Data-centric approach

 graph TD;
    A["Input"] -->|Synchronous| B["Routing Engine"];
    A2["Input"] -->|Synchronous| B;
    
    B --> C["Custom Logic"];
    
    C --> D["ML Model B"];
    C --> E["ML Model A"];
    C -->|Synchronous| F["Drift Detector"];
    
    D --> G["ML Model C (LLM)"];
    G --> H["Output"];
    
    F -->|Asynchronous| I["Drift Alerts"];

    style F stroke-dasharray: 5,5;
    style I stroke-dasharray: 5,5;
    
    subgraph Legend
        J["Synchronous"]
        K["Asynchronous"]:::dashed
    end
    
    classDef dashed stroke-dasharray: 5,5;


{% embed url="https://www.youtube.com/watch?v=ar5lSG_idh4" %}

Seldon Core 2 offers a powerful, modular framework that enables businesses to deploy, monitor, and optimize ML models with key benefits such as:

## Flexibility: Real-time, your way

Seldon Core 2 is designed to accommodate diverse machine learning deployment needs, offering a platform- and integration-agnostic framework that ensures seamless deployment across on-premise, cloud, and hybrid environments. Unlike rigid infrastructure solutions, the flexible architecture of Seldon Core 2 empowers businesses to build highly customizable applications tailored to their unique business and technical requirements.
The adaptive framework future-proofs MLOps investments, enabling organizations to scale their ML deployments as applications and data evolve, whether in staging, testing, or production environments. Additionally, its modular design optimizes cost-efficiency and sustainability by allowing businesses to scale ML systems up or down as needed, reuse components across workflows, and maximize existing infrastructure investments.
This approach helps organizations dynamically respond to changing demands, deploy resources efficiently, and maintain flexibility across any deployment environment, ensuring long-term scalability and operational efficiency.

## Standardization: consistency across workflows

Seldon Core 2 enforces industry best practices for ML deployment, ensuring consistency, reliability, and efficiency across the entire machine learning lifecycle. By creating repeatable processes, organizations can deploy models faster, course correct efficiently, and gain a competitive advantage. The platform automates critical deployment steps, freeing teams from operational bottlenecks so they can focus on high-value tasks instead of repetitive workflows.

With a "learn it once, repeat everywhere" approach, Seldon Core 2 ensures that the same streamlined, standardized process can be used to deploy models anywhere, whether on-premise, in the cloud, or hybrid environments. This scalability and efficiency reduce deployment risks and enhance overall productivity. Furthermore, Seldon’s standardized methods build trust in both existing and newly deployed models, no matter how many are in use. This consistency encourages innovation and unlocks the unrealized potential of AI, allowing businesses to confidently scale their ML operations while maintaining accuracy and compliance.

## Enhanced Observability: Data-Centric monitoring

Observability in Seldon Core 2 provides the ability to monitor, understand, and analyze the behavior, performance, and health of an ML system across its entire lifecycle—including data pipelines, models, and deployment environments. By offering a customizable observability framework, Seldon Core 2 seamlessly combines operational monitoring and data science monitoring, ensuring that teams have access to key metrics necessary for both model maintenance and strategic decision-making.

With scalable maintenance for complex applications, Seldon simplifies operational monitoring while enabling teams to expand real-time ML deployments across the organization—supporting increasingly sophisticated and mission-critical use cases. A data-centric oversight approach ensures that all data involved in predictions can be reviewed and audited, enabling organizations to maintain explainability, compliance, and trust in AI-driven decisions.

Seldon Core 2 also features flexible metric aggregation, surfacing operational, data science, and custom metrics tailored to different user personas and specific needs. Whether teams require high-level overviews or granular insights, Seldon ensures transparency at every level. With real-time access and insights, users can register, monitor, and audit models step by step, gaining full visibility into what data influences predictions and how decisions are made—allowing organizations to build more accountable, trustworthy AI systems.

## Optimization: Modularity for efficient scaling

Seldon Core 2 is designed to maximize resource efficiency, reduce infrastructure costs, and enhance collaboration while ensuring scalability and high-performance ML operations. Its modular architecture enables businesses to deploy only the necessary components, avoiding unnecessary expenses while maintaining agility and efficiency.

* Optimized resource utilization: Seldon allows users to scale infrastructure dynamically based on real-time demand, ensuring that only required resources are used. By eliminating redundancy and reusing models across workflows, organizations can reduce operational costs while maintaining efficiency.

* Streamlined development: The flexible, standardized, and modular approach of Seldon enables you to build, adapt, and repurpose models efficiently, eliminating the need to start from scratch. This accelerates ML development, enhances operational efficiency, and optimizes time and resource management.

* Seamless collaboration: Seldon Core 2 fosters better communication and integration between MLOps Engineers, Data Scientists, and Software Engineers. By providing a customizable framework, it supports knowledge sharing, encourages innovation, and simplifies the adoption of new data science-focused features.

* Sustainable AI scaling: By optimizing infrastructure usage and reducing the need for continuous re-engineering, Seldon Core 2 helps organizations minimize computational expenses while ensuring long-term AI sustainability. Its modular components can be tailored and repurposed, making AI deployments more cost-effective and adaptable to evolving business needs.

By combining scalability, modular efficiency, and collaborative innovation, Seldon Core 2 empowers you to lower costs, streamline ML workflows, and drive AI-driven innovation—all while maintaining flexibility and high performance. 

## Next Steps

- [Install Seldon Core 2](./getting-started/README.md)
- Explore our [Tutorials](./examples/README.md)
- [Join our Slack Community](https://seldondev.slack.com/join/shared_invite/zt-vejg6ttd-ksZiQs3O_HOtPQsen_labg#/shared-invite/email) for updates or for answers to any questions
