<p align="center">
    <img alt="Finn Logo" src="/static/logo.png?v=0.1.0" width="180" />
    <h3 align="center">Finn</h3>
    <p align="center">A Platform to Run, Schedule, and Automate Ansible and Terraform Workflows Powered With AI.</p>
    <p align="center">
        <a href="https://github.com/Clivern/Finn/actions/workflows/api.yml">
            <img src="https://github.com/Clivern/Finn/actions/workflows/api.yml/badge.svg">
        </a>
        <a href="https://github.com/Clivern/Finn/releases">
            <img src="https://img.shields.io/badge/Version-v0.1.0-red.svg">
        </a>
        <a href="https://github.com/Clivern/Finn/blob/main/LICENSE">
            <img src="https://img.shields.io/badge/LICENSE-MIT-grey.svg">
        </a>
    </p>
</p>

Finn is a modern platform built for DevOps teams for running, scheduling, and automating `Ansible` and `Terraform` workflows. Run playbooks and Terraform workflows from a web interface or API without direct CLI access, and keep infrastructure and config management under one platform.

**Features:**

- **Run automation from the UI** — Execute Ansible playbooks and Terraform plans/applies via a web interface or a full-featured API (e.g. OpenAPI/Swagger) for integration with CI/CD and other tools
- **Schedule and automate** — Schedule tasks for recurring runs (drift checks, updates, backups) or trigger executions on demand
- **Manage inventories & state** — Centrally manage Ansible inventories, host groups, and variables; run Terraform with consistent state and credentials
- **Terraform HTTP backend with state locking** — HTTP backend for Terraform state with locking support, so you can store and lock state centrally without external backends
- **Monitor execution** — View logs, execution status, and errors in one place for troubleshooting
- **Control access & secrets** — Control who can run tasks and modify config; store and use SSH keys, API tokens, and other secrets securely
- **Single execution layer** — Run both Ansible and Terraform (and combined workflows) from one platform with consistent logging and visibility


### Versioning

For transparency into our release cycle and in striving to maintain backward compatibility, Finn is maintained under the [Semantic Versioning guidelines](https://semver.org/) and release process is predictable and business-friendly.

See the [Releases section of our GitHub project](https://github.com/Clivern/Finn/releases) for changelogs for each release version of Finn. It contains summaries of the most noteworthy changes made in each release. Also see the [Milestones section](https://github.com/Clivern/Finn/milestones) for the future roadmap.


### Bug tracker

If you have any suggestions, bug reports, or annoyances please report them to our issue tracker at https://github.com/Clivern/Finn/issues


### Security Issues

If you discover a security vulnerability within Finn, please send an email to [hello@clivern.com](mailto:hello@clivern.com)


### Contributing

We are an open source, community-driven project so please feel free to join us. see the [contributing guidelines](CONTRIBUTING.md) for more details.


### License

© 2025, Clivern. Released under [MIT License](https://opensource.org/licenses/mit-license.php).

**Finn** is authored and maintained by [@Clivern](http://github.com/Clivern).
