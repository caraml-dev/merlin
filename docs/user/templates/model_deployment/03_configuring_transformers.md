<!-- page-title: Configuring Transformers -->
<!-- parent-page-title: Deploying a Model -->
# Transformer

In the Merlin ecosystem, a Transformer is a service deployed in front of the model service which users can use to perform pre-processing / post-processing steps to the incoming request / outgoing response, to / from the model service. A Transformer allows the user to abstract the transformation logic outside of their model and even write it in a language more performant than python.

Currently, Merlin supports two types of Transformer: Standard and Custom:

{% page-ref page="./transformer/01_standard_transformer.md" %}

{% page-ref page="./transformer/02_custom_transformer.md" %}