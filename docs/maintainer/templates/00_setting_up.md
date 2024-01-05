<!-- page-title: Setting Up Merlin -->
# Installing Merlin

Merlin can be installed using the Helm charts located at [caraml-dev/helm-charts](https://github.com/caraml-dev/helm-charts/tree/main).

Minimally, [MLP](https://github.com/caraml-dev/mlp) and [KServe](https://github.com/kserve/kserve) must be installed for Merlin to work. Besides these, a production deployment of Merlin would require other components such as networking, authorization policies, etc. to be set up. All of these capabilities are provided by the umbrella [CaraML chart](https://github.com/caraml-dev/helm-charts/tree/main/charts/caraml). It is recommended to install this chart using the appropriate toggles and configurations for its different sub-components.

# Configuring Merlin

Besides the configurations documented by the CaraML umbrella chart, detailed specs may be found under each of the sub-charts. For example, the [Merlin chart](https://github.com/caraml-dev/helm-charts/tree/main/charts/merlin)'s docs capture the list of configurable parameters. Additional configurations (`config.*`) accepted by Merlin may also be found [here](https://github.com/caraml-dev/merlin/blob/main/api/config/config.go#L46).
