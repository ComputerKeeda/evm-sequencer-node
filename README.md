![Project Logo](https://www.airchains.io/assets/logos/airchains-evm-rollup-full-logo.png)

# Overview

EVM Chain Sequencer is a high-performance, innovative tool designed to optimize transaction and block management on Ethereum Virtual Machine (EVM) chains. This tool stands out for its integration of advanced batching and Data Availability (DA) processes, ensuring efficient and reliable blockchain operations.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Key Features](#key-features)
- [Usage](#usage)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Key Features

- **Enhanced Transaction Batching**: Implements sophisticated algorithms for efficient transaction aggregation, significantly improving throughput and reducing latency.

- **Data Availability (DA) Processes**: Incorporates robust DA mechanisms to ensure data integrity and accessibility, enhancing trust and transparency in the blockchain network.

- **Seamless Settlement Layer Integration**: Designed for smooth interaction with the settlement layer, maintaining operational integrity and consistent performance.

- **High Throughput and Reliability**: Focuses on handling large volumes of transactions effectively, ensuring both high throughput and steadfast reliability in blockchain operations.

## Usage

In order to tailor the Sequencer to better align with your specific requirements, please proceed to update key configuration parameters within the `common/constants.go` file. The following constants are crucial for the optimal functioning of the sequencer and can be adjusted to meet your operational needs:

- **BatchSize**: Modify this value to alter the batch size for transaction processing. This adjustment can optimize throughput and efficiency based on your workload.

- **BlockDelay**: Adjust this constant to set the delay between blocks check, aligning it with your network's block generation rate for synchronized operations.

- **ExecutionClientRPC**: Update this URL to connect the sequencer with your preferred execution client's RPC interface.

- **SettlementClientRPC**: Change this URL to integrate the sequencer with the desired settlement layer's RPC service.

- **KeyringDirectory**: Specify a new directory path for the keyring, ensuring secure and organized storage of cryptographic keys.

- **DaClientRPC**: Alter this URL to link the sequencer with your chosen Data Availability (DA) service's RPC endpoint.

Each of these parameters plays a critical role in the configuration and performance of the sequencer. It is recommended to carefully consider the implications of these changes to maintain optimal functionality and security of the system.

> Note: before proceeding to run the sequencer, please ensure that the `init_dir.sh` script has been executed to initialize the basic directory structure and configuration files.

_Important Security Notice Regarding init_dir.sh Execution_

Please be aware that running the `init_dir.sh` script necessitates the entry of your terminal password. This requirement stems from the inclusion of `sudo` commands within the script. These commands elevate privileges for certain operations, which are essential for the correct setup and configuration of the environment.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

Special thanks to the `gnark` library, an efficient and elegant toolkit for zk-SNARKs on Go. This library has been instrumental in our development process. For more information and to explore their work, visit their GitHub repository at [Consensys/gnark.](https://github.com/Consensys/gnark)
