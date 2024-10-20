<a id="readme-top"></a>

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

  <h3 align="center">GoCrypt</h3>

  <p align="center">
    A decentralized File Storage System written in Go!
    <br />
    <a href="https://github.com/its-kos/gocrypt">View Demo</a>
    ·
    <a href="https://github.com/its-kos/gocrypt/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    ·
    <a href="https://github.com/its-kos/gocrypt/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
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
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>


<!-- ABOUT THE PROJECT -->
## About The Project

This is a fun hobby project I created to learn Go more in depth. This tool splits the file you input, encrypts then communicated the chunks to other peers in a P2P network. When acces is needed, peers communicate the requested chunks, decrypt them, the stitch the original file together. 

The original scope for this is to be used for a P2P network of nodes in the same network for simple files. Individual setup on each node is needed.


<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

Go 1.23+ is required in order to build the project.

### Installation

1. `go build .` in the project directory
2. Make a "files/testfile.ext" directory. The fletype can be whatever.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

1. `./gocrypt -host <local-network-ip> -port <free-tcp-port>` on two different terminals.
2. (For now the test file you input is automatically chunked and uploaded)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

- [x] Basic chunking functionality
- [x] Basic encrypting functionality
- [x] libp2p local network setup
- [ ] finish documentation
- [ ] cli implementation
    - [x] basic boilerplate / scaffolding
    - [ ] command implementation

See the [open issues](https://github.com/its-kos/gocrypt) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. Or you can also simply open an issue.
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Top contributors:

<a href="https://github.com/its-kos/gocrypt/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=its-kos/gocrypt" alt="contrib.rocks image" />
</a>

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- LICENSE -->
## License

No License as of now. (Don't) See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Konstantinos Katserelis - [@gravityWell](https://twitter.com/gravityWwell)

Project Link: [https://github.com/its-kos/gocrypt](https://github.com/its-kos/gocrypt)


<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/its-kos/gocrypt.svg?style=for-the-badge
[contributors-url]: https://github.com/its-kos/gocrypt/graphs/contributors

[forks-shield]: https://img.shields.io/github/forks/its-kos/gocrypt.svg?style=for-the-badge
[forks-url]: https://github.com/its-kos/gocrypt/network/members

[stars-shield]: https://img.shields.io/github/stars/its-kos/gocrypt.svg?style=for-the-badge
[stars-url]: https://github.com/its-kos/gocrypt/stargazers

[issues-shield]: https://img.shields.io/github/issues/its-kos/gocrypt.svg?style=for-the-badge
[issues-url]: https://github.com/its-kos/gocrypt/issues

[license-shield]: https://img.shields.io/github/license/its-kos/gocrypt.svg?style=for-the-badge
[license-url]: https://github.com/its-kos/gocrypt/blob/master/LICENSE.txt

[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/konstantinos-katserelis/

[Golang]: https://img.shields.io/badge/go-000000?style=for-the-badge&logo=go
[Go-url]: https://go.dev/