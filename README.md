# Go Streamlink Integration for Jellyfin

## Overview

This project provides a Go application designed to interface between live streams captured via Streamlink and the Jellyfin media server. The application generates an M3U playlist dynamically, making it possible for Jellyfin to access and display live streams as part of its media offerings. This setup is ideal for users looking to incorporate live streaming content directly into their Jellyfin server environment.

## Features

- **Dynamic M3U Playlist Generation**: Automatically generates and updates playlists for live streaming.
- **Streamlink Integration**: Leverages Streamlink to fetch and serve live streams.
- **Easy Integration with Jellyfin**: Designed to be integrated with Jellyfin for streaming live content.
- **Docker Support**: Includes Dockerfile for easy deployment.

## Prerequisites

- Go 1.16 or higher
- Docker (for containerization)
- Jellyfin Media Server
- Streamlink

## Installation

Here's what worked for me. Note that I wrote this so that'll work for my use case. Might need to change a thing here and there:

### Clone the Repository

First, clone this repository to your local machine:
```bash
    git clone https://github.com/yourusername/yourrepository.git
    cd yourrepository
```
### Build the Application

Compile the Go application:
```bash
    go build -o streamlink-service
```
### Build Docker Image

If you prefer to run the application in a Docker container:
```bash
    docker build -t streamlink-service .
```

## Running the Application

To run the application directly:

    ./streamlink-service

To run using Docker:

    docker run -d -p 8801:8801 streamlink-service

## Integration with Jellyfin

To integrate with Jellyfin:

 **Add a Live TV M3U Tuner**:
   - Go to Jellyfin's Dashboard.
   - Navigate to Live TV under the settings.
   - Add a new tuner, choose M3U Tuner, and provide the URL where this service is running, e.g., `http://yourserverip:8801/`.

## Usage

Once the service is running and configured in Jellyfin, you should be able to view the live streams under the Live TV section of Jellyfin.

## Contributing

Make a fork. I will not maintain this repo

## License

This project is licensed under the MIT License - see the LICENSE.md file for details.
