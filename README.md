<p align="center">

<a href="https://docs.dadosfera.ai">
  <img src="https://files.readme.io/2b7adf2-Logo_Dadosfera_V3_Logo_without_slogan.png" width="350px" />
</a>
</p>

<p align="center">
<a href=https://docs.dadosfera.ai><img src="https://img.shields.io/badge/Website-blue?style=flat&logo=webflow&labelColor=5c5c5c"></a>
<a href=https://docs.dadosfera.ai><img src="https://img.shields.io/badge/Dadosfera SaaS-blue?style=flat&logo=aws&labelColor=5c5c5c"></a>
<a href=https://docs.dadosfera.ai><img src="https://readthedocs.org/projects/orchest/badge/?version=stable&style=flat"></a>
<a href=https://docs.dadosfera.ai><img src="https://img.shields.io/badge/Video%20tutorials-blue?style=flat&logo=airplayvideo&labelColor=5c5c5c"></a>
<a href=https://docs.dadosfera.ai><img src="https://img.shields.io/badge/Quickstart-blue?style=flat&logo=readthedocs&labelColor=5c5c5c&color=fc0373"></a>
<a href=https://discord.gg/CHpc6K3vZp><img src="https://img.shields.io/badge/Discord-violet?style=flat&logo=discord&labelColor=5c5c5c"></a>
</p>

## Dadosfera Fork

This project is a private fork of the orchest project https://github.com/orchest/orchest.
It allows modifying the services: orchest-webserver and auth-server that are inside the services folder.
When modifying this project in the main branch, a new version is generated and the containers are built by github actions.
To deploy the new version on EKS clusters, use the orchest-eks project https://github.com/dadosfera/orchest-eks.

### Development process

1. Develop, commit and push commits to remote main;
2. An action will run and it will return the [release vesion number](https://github.com/dadosfera/orchest/releases) ;
3. Copy the release version value and paste it into the [orchest-cluster.yml](https://github.com/dadosfera/orchest-eks-deploy/blob/main/files/orchest-cluster.yml) file in the [orchest-eks-deploy](https://github.com/dadosfera/orchest-eks-deploy) repo replacing at the end of the 'image' field;
4. Commit and push the repo orchest-eks-deploy to remote repo;
5. On [Github Actions orchest eks deploy](https://github.com/dadosfera/orchest-eks-deploy/actions/workflows/deploy.yml), choose the desired environment (stg or prd) and trigger to send it;
6. Each customer has its own env, to find out which customers are active, go to [customers.tf](https://github.com/dadosfera/terraform-ddosfera/blob/main/prd_611330257153/customers.tf), and you would see some of `create_customer_orchest_intelligence` and `create_customer_orchest_process` boolean properties, if any of these is true, then it has orchest, otherwise if false or undefined, it doesn't.

### Observations:

- The preferred NPM version need to be the same as defined in [package.json](https://github.com/dadosfera/orchest/blob/master/package.json) at `engine` object

## Update Orchest Version

Follow tutorial to update private fork through orchest source code:

```bash
git remote add public https://github.com/orchest/orchest.git
git pull public master
git push origin master
```

If any conflicts occur, resolve the conflicts and push.
Update orchest version in .github/workflows/deploy.yml and .releaserc.json.

## Build data pipelines, the easy way 🙌

No frameworks. No YAML. Just write your data processing code directly in **Python**, **R** or
**Julia**.

<p align="center">
  <img width="100%" src="https://user-images.githubusercontent.com/1309307/191785568-ce4857c3-e71f-4b71-84ce-dfa5d65a98f9.gif">
</p>

<p align="center">
  <i>💡 Watch the <a target="_blank" href="https://vimeo.com/764866337">full narrated video</a> to learn more about building data pipelines in Orchest.</i>
 </p>

> **Note**: Orchest is in **beta**.

## Features

- **Visually construct pipelines** through our user-friendly UI
- **Code in Notebooks** and scripts
  ([quickstart](https://docs.orchest.io/en/stable/getting_started/quickstart.html))
- Run any subset of a pipelines directly or periodically
  ([jobs](https://docs.orchest.io/en/stable/fundamentals/jobs.html))
- Easily define your dependencies to run on **any machine**
  ([environments](https://docs.orchest.io/en/stable/fundamentals/environments.html))
- Spin up services whose lifetime spans across the entire pipeline run
  ([services](https://docs.orchest.io/en/stable/fundamentals/services.html))
- Version your projects using git
  ([projects](https://docs.orchest.io/en/stable/fundamentals/projects.html))

**When to use Orchest?** Read it in the
[docs](https://docs.orchest.io/en/stable/getting_started/when_to_use_orchest.html).

👉 Get started with our
[quickstart](https://docs.orchest.io/en/stable/getting_started/quickstart.html) tutorial or have a look at our [video tutorials](https://www.orchest.io/video-tutorials) explaining some of Orchest's core concepts.

## Roadmap

Missing a feature? Have a look at [our public roadmap](https://github.com/orgs/orchest/projects/1)
to see what the team is working on in the short and medium term.
Still missing it? Please [let us know by opening an issue](https://github.com/orchest/orchest/issues/new/choose)!

## Examples

Get started with an example project:

- [Train and compare 3 regression models](https://github.com/orchest/quickstart)
- [Connecting to an external database using SQLAlchemy](https://github.com/astrojuanlu/orchest-sqlalchemy)
- [Run dbt in Orchest for a dbt + Python transform pipeline](https://github.com/ricklamers/orchest-dbt)
- [Use PySpark in Orchest](https://github.com/ricklamers/orchest-hello-spark)

👉 Check out the full list of [example projects](https://github.com/orchest/orchest-examples).

[![Open in Orchest](https://github.com/orchest/orchest-examples/raw/main/imgs/open_in_orchest_large.svg)](https://cloud.orchest.io/)

## Installation

Want to skip [the installation](https://docs.orchest.io/en/stable/getting_started/installation.html)
and jump right in? Then try out our managed service: [Dadosfera AI Cloud](https://dadosfera.ai).

## Development

To run Orchest locally for development, follow our detailed [development workflow guide](docs/source/development/development_workflow.md). This guide includes:

- Required prerequisites
- Development environment setup
- How to run Orchest locally
- How to make changes and test
- How to contribute to the project

### Notes for Intel Mac Users

If you are using a Mac with an Intel processor, some additional configurations may be required:

1. Make sure Docker Desktop is configured to use the "hyperkit" driver instead of "qemu"
2. For minikube, use the "hyperkit" driver:
```bash
minikube start --driver=hyperkit
```
3. If you encounter MySQL issues, install via Homebrew:
```bash
brew install mysql
```

For more details about Intel Mac specific configuration, check the [prerequisites section](docs/source/development/development_workflow.md#prerequisites) in the development guide.

## Discord Community

Join our Discord to chat about Orchest, ask questions, and share tips.

[![Join us on Discord](https://img.shields.io/badge/%20-Join%20us%20on%20Discord-blue?style=for-the-badge&logo=discord&labelColor=5c5c5c)](https://discord.gg/CHpc6K3vZp)

## License

This software is licensed under the Elastic License 2.0 (ELv2). See the [LICENSE](LICENSE) file for details.

The Elastic License 2.0 (ELv2) is a permissive license that allows you to:
- Use the software for any purpose
- Modify the software
- Distribute the software
- Use the software commercially

However, you must:
- Include the original copyright notice
- Include the license text
- State significant changes made to the software
- Include the Elastic License 2.0 (ELv2) license

For more information about the Elastic License 2.0, please visit: https://www.elastic.co/licensing/elastic-license

## Contributing

Contributions are more than welcome! Please see our [contributor
guides](https://docs.orchest.io/en/stable/development/contributing.html) for more details.

Alternatively, you can submit your pipeline to the curated list of [Orchest
examples](https://github.com/orchest/orchest-examples) that are automatically loaded in every
Orchest deployment! 🔥

## Contributors

<!-- To get src for img: https://api.github.com/users/username -->
<a href="https://github.com/allansene"><img src="https://avatars.githubusercontent.com/allansene?v=4" title="Allan Sene" width="50" height="50"></a>
<a href="https://github.com/rafaelsantanaep"><img src="https://avatars.githubusercontent.com/rafaelsantanaep?v=4" title="Rafael Santana" width="50" height="50"></a>
<a href="https://github.com/ricklamers"><img src="https://avatars2.githubusercontent.com/u/1309307?v=4" title="ricklamers" width="50" height="50"></a>
<a href="https://github.com/yannickperrenet"><img src="https://avatars0.githubusercontent.com/u/26223174?v=4" title="yannickperrenet" width="50" height="50"></a>
<a href="https://github.com/fruttasecca"><img src="https://avatars3.githubusercontent.com/u/19429509?v=4" title="fruttasecca" width="50" height="50"></a>
<a href="https://github.com/samkovaly"><img src="https://avatars2.githubusercontent.com/u/32314099?v=4" title="samkovaly" width="50" height="50"></a>
<a href="https://github.com/VivanVatsa"><img src="https://avatars0.githubusercontent.com/u/56357691?v=4" title="VivanVatsa" width="50" height="50"></a>
<a href="https://github.com/obulat"><img src="https://avatars1.githubusercontent.com/u/15233243?v=4" title="obulat" width="50" height="50"></a>
<a href="https://github.com/howie6879"><img src="https://avatars.githubusercontent.com/u/17047388?v=4" title="howie6879" width="50" height="50"></a>
<a href="https://github.com/FanaHOVA"><img src="https://avatars.githubusercontent.com/u/6490430?v=4" title="FanaHOVA" width="50" height="50"></a>
<a href="https://github.com/mitchglass97"><img src="https://avatars.githubusercontent.com/u/52224377?v=4" title="mitchglass97" width="50" height="50"></a>
<a href="https://github.com/joe-bell"><img src="https://avatars.githubusercontent.com/u/7349341?v=4" title="joe-bell" width="50" height="50"></a>
<a href="https://github.com/cceyda"><img src="https://avatars.githubusercontent.com/u/15624271?v=4" title="cceyda" width="50" height="50"></a>
<a href="https://github.com/MWeltevrede"><img src="https://avatars.githubusercontent.com/u/31962715?v=4" title="MWeltevrede" width="50" height="50"></a>
<a href="https://github.com/kingabzpro"><img src="https://avatars.githubusercontent.com/u/36753484?v=4" title="Abid" width="50" height="50"></a>
<a href="https://github.com/iannbing"><img src="https://avatars.githubusercontent.com/u/627607?v=4" title="iannbing" width="50" height="50"></a>
<a href="https://github.com/andtheWings"><img src="https://avatars.githubusercontent.com/u/5892089?v=4" title="andtheWings" width="50" height="50"></a>
<a href="https://github.com/jacobodeharo"><img src="https://avatars.githubusercontent.com/jacobodeharo?v=4" title="jacobodeharo" width="50" height="50"></a>
<a href="https://github.com/nhaghighat"><img src="https://avatars.githubusercontent.com/u/3792293?v=4" title="nhaghighat" width="50" height="50"></a>
<a href="https://github.com/porcupineyhairs"><img src="https://avatars.githubusercontent.com/u/61983466?v=4" title="porcupineyhairs" width="50" height="50"></a>
<a href="https://github.com/ncspost"><img src="https://avatars.githubusercontent.com/ncspost?v=4" title="ncspost" width="50" height="50"></a>
<a href="https://github.com/cavriends"><img src="https://avatars.githubusercontent.com/u/4497501?v=4" title="cavriends" width="50" height="50"></a>
<a href="https://github.com/astrojuanlu"><img src="https://avatars.githubusercontent.com/u/316517?v=4" title="astrojuanlu" width="50" height="50"></a>
<a href="https://github.com/mausworks"><img src="https://avatars.githubusercontent.com/u/8259221?v=4" title="mausworks" width="50" height="50"></a>
<a href="https://github.com/jerdna-regeiz"><img src="https://avatars.githubusercontent.com/u/7195718?v=4" title="jerdna-regeiz" width="50" height="50"></a>
<a href="https://github.com/sbarrios93"><img src="https://avatars.githubusercontent.com/u/19554889?v=4" title="sbarrios93" width="50" height="50"></a>
<a href="https://github.com/cacrespo"><img src="https://avatars.githubusercontent.com/u/10950697?v=4" title="cacrespo" width="50" height="50"></a>
