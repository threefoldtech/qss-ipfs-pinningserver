<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a name="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/threefoldtech/tf-pinning-service">
    <img src="https://library.threefold.me/info/threefold/threefold__peoples_internet_.png" alt="Logo" width="100%">
  </a>

<h1 align="center">TF Pinning Service</h3>

  <p align="center">
    IPFS Pinning Service over TF Quantum-Safe Storage
    <br />
    <a href="https://github.com/threefoldtech/tf-pinning-service"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="#">View Demo</a>
    ·
    <a href="https://github.com/threefoldtech/tf-pinning-service/issues">Report Bug</a>
    ·
    <a href="https://github.com/threefoldtech/tf-pinning-service/issues">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#compile">Compile</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <ul>
        <li><a href="#environment-variables">Environment Variables</a></li>
        <li><a href="#docker">Docker</a></li>
        <li><a href="#docker-compose">Docker Compose</a></li>
    </ul>
    <li><a href="#interacting-with-the-pinning-service">Interacting With The Pinning Service</a></li>
    <ul>
        <li><a href="#using-the-http-api">Using the HTTP API</a></li>
        <li><a href="#using-the-ipfs-cli">Using the IPFS CLI</a></li>
        <li><a href="#using-the-ipfs-desktop-gui-app">Using the IPFS Desktop GUI app</a></li>
    </ul>
    <li><a href="#api-specs">API Specs</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project


IPFS pinning service backed by ThreeFold decentralized grid and Quantum Safe Filesystem.


<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [Go](https://go.dev/)
* [IPFS](https://ipfs.io/)
* [QSFS](https://github.com/threefoldtech/0-stor_v2)
* [Vue](https://vuejs.org/)




<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Threefold pinning service is a IPFS pinning service that complies with the IPFS Pinning Service API specification. it combines the content addressing capabilities of IPFS, the decentralized, Peer To Peer Internet infrastructure of Threefold, and Threefold's distributed and ultra-secure storage backend QSFS(Quantum-Safe Storage Filesystem) in order to achieve a powerful decentralized storage solution, wrapping the power of these systems into an easy-to-use service with simple rest API.

### Prerequisites

- An IPFS cluster with at least one IPFS peer.
  - For development you can spin up a local IPFS Cluster instance. [setup instructions](https://docs.ipfs.tech/install/server-infrastructure/#create-a-local-cluster).
  - For setup IPFS and IPFS Cluster on a production environment see [here](https://ipfscluster.io/documentation/deployment/)
- The tfpin web service binary.
  - You can download it from releases or compile from the source code.
  If you need to compile from the source code, you will need also:
    - An installation of Go 1.16 or later. [installation instructions](https://go.dev/doc/install)
    - Git client


### Compile

To compile tfpin binary from the source code, follow below instructions 

1 - Clone the repository:
  - Open Terminal, Change the current working directory to the location where you want the cloned directory, and type
  ```sh
  git clone https://github.com/threefoldtech/tf-pinning-service.git
  ```
2- Compile:
  - Change to `./tf-pinning-service` and type
  ```sh
  make build
  ```

  Then find the compiled binary file `tfpin` in the repo root directory.

Other make scripts available.

```sh
make run        # run the web server
```

```sh
make build_run  # compile and run the web server
```

Make sure to set correctly the required environment variables. see below usage section.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

1- set the required environment variables (see next section), then run the compiled binary

```sh
TFPIN_CLUSTER_HOSTNAME="cluster-host-name" TFPIN_CLUSTER_USERNAME="usname" TFPIN_CLUSTER_PASSWORD="password" ./tfpin
```

2- You will need API token to interact with the service, for development run the script below to add a token of your choice (the development of the real sign-up flow blocked on the the chain part for now):

```sh
go run ./scripts/add_test_tokens.go BestTokenEver
Token `BestTokenEver` stored successfully.
```

### Environment Variables

| Environment Variable  	| Description  	| Default Value  	|
|---	|---	|---	|
| TFPIN_CLUSTER_HOSTNAME  	|   	| 127.0.0.1  	|
| TFPIN_CLUSTER_PORT  	|   	| 9097  	|
| TFPIN_CLUSTER_USERNAME  	|   	| ""  	|
| TFPIN_CLUSTER_PASSWORD  	|   	| ""  	|
| TFPIN_CLUSTER_REPLICA_MIN  	|   	| -1 (Pin on all cluster IPFS peers) 	|
| TFPIN_CLUSTER_REPLICA_MAX  	|   	| -1 (Pin on all cluster IPFS peers) 	|
| TFPIN_DB_DSN  	|   	| pins.db  	|
| TFPIN_DB_LOG_LEVEL  	|   	| 1 (silent) 	|
| TFPIN_SERVER_ADDR  	|   	| :8000  	|
| TFPIN_SERVER_LOG_LEVEL  	|   	| 3 (warn) 	|
| TFPIN_AUTH_HEADER_KEY  	|   	| Authorization  	|

### Docker
#### Building Image

```sh
docker build -t abouelsaad/tfpinsvc .
```

#### Running the container

```sh
docker run --name tfpinsvc --env TFPIN_CLUSTER_HOSTNAME=HOSTNAME --env TFPIN_CLUSTER_PORT=9094 TFPIN_CLUSTER_USERNAME=USERNAME --env TFPIN_CLUSTER_PASSWORD=PASSWORD --env TFPIN_SERVER_ADDR=:80 -p 8000:80 abouelsaad:tfpin
```

Adding test token

```sh
docker exec tfpinsvc ./add_test_tokens BestTokenEver
```
### Docker Compose

For development and testing, with one command you can spin up
- an IPFS cluster, and
- Threefold pinning service

which is already configured to communicate, so you just need to get/add a test token, and start interacting with the service.

```sh
docker compose up -d
docker exec tfpinsvc ./add_test_tokens BestTokenEver
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- Interacting With The Pinning Service -->
## Interacting With The Pinning Service

To interact with the pinning service you can use any of:
- IPFS Desktop or IPFS Web UI [installation instructions](https://github.com/ipfs/ipfs-desktop#install)
- IPFS CLI [installation instructions](https://docs.ipfs.tech/install/command-line/)
- Any http client, like `curl`

See Use pinning service [instructions](https://docs.ipfs.tech/how-to/work-with-pinning-services/#use-an-existing-pinning-service)

The threefold pinning service endpoint for all requests is
https://[hostname]/api/v1/pins

### Using the HTTP API
#### Add a pin

```sh
curl -X POST 'https://[HOSTNAME]/api/v1/pins' \
  --header 'Accept: */*' \
  --header 'Authorization: Bearer <YOUR_AUTH_TOKEN>' \
  --header 'Content-Type: application/json' \
  -d '{
  "cid": "<CID_TO_BE_PINNED>",
  "name": "PreciousData.pdf"
}'

```

#### List successful pins

```sh
curl -X GET 'https://[HOSTNAME]/api/v1/pins' \
  --header 'Accept: */*' \
  --header 'Authorization: Bearer <YOUR_AUTH_TOKEN>'
```

#### Delete a pin

```sh
curl -X DELETE 'https://[HOSTNAME]/api/v1/pins/<REQUEST_ID>' \
  --header 'Accept: */*' \
  --header 'Authorization: Bearer <YOUR_AUTH_TOKEN>'
```

### Using the IPFS CLI
The IPFS CLI can be used to maintain pins by first adding the threefold pinning service.

```sh
ipfs pin remote service add tfpinsvc https://[HOSTNAME]/api/v1/ <YOUR_AUTH_TOKEN>
```

#### Add a pin

```sh
ipfs pin remote add --service=tfpinsvc --name=<PIN_NAME> <CID>
```
#### List successful pins

```sh
ipfs pin remote ls --service=tfpinsvc
```
#### Delete a pin

```sh
ipfs pin remote rm --service=tfpinsvc --cid=<CID>
```

### Using the IPFS Desktop GUI app

see [here](https://docs.ipfs.tech/how-to/work-with-pinning-services/#ipfs-desktop-or-ipfs-web-ui)

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- API SPECS -->
## API Specs

Threefold pinning service is compatible with the IPFS Pinning Service API (1.0.0) OpenAPI spec. 
see [here](./pinning-api/README.md)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] simple ipfs pining service server compatible with this OpenAPI spec (https://github.com/ipfs/pinning-services-api-spec).  
- [ ] registration using smart contract
- [ ] payment
    

See the [open issues](https://github.com/threefoldtech/tf-pinning-service/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the Apache License. See [`LICENSE`](LICENSE) for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Threefold - [@threefold_io](https://twitter.com/threefold_io)

Project Link: [https://github.com/threefoldtech/tf-pinning-service](https://github.com/threefoldtech/tf-pinning-service)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [Pinning services api spec](https://ipfs.github.io/pinning-services-api-spec/)
* [Work with pinning services](https://docs.ipfs.tech/how-to/work-with-pinning-services/#use-an-existing-pinning-service)
* [Threefold quantum safe filesystem](https://library.threefold.me/info/manual/#/technology/threefold__qsfs)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/threefoldtech/tf-pinning-service.svg?style=for-the-badge
[contributors-url]: https://github.com/threefoldtech/tf-pinning-service/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/threefoldtech/tf-pinning-service.svg?style=for-the-badge
[forks-url]: https://github.com/threefoldtech/tf-pinning-service/network/members
[stars-shield]: https://img.shields.io/github/stars/threefoldtech/tf-pinning-service.svg?style=for-the-badge
[stars-url]: https://github.com/threefoldtech/tf-pinning-service/stargazers
[issues-shield]: https://img.shields.io/github/issues/threefoldtech/tf-pinning-service.svg?style=for-the-badge
[issues-url]: https://github.com/threefoldtech/tf-pinning-service/issues
[license-shield]: https://img.shields.io/github/license/threefoldtech/tf-pinning-service.svg?style=for-the-badge
[license-url]: https://github.com/threefoldtech/tf-pinning-service/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/linkedin_username
