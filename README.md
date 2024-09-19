# Sales Data Visualization App

This application visualizes sales data over time using a line chart. It is built with Go for the backend and Chart.js for the frontend.

## Features

- Ingest sales data (from CSV file)
- Store sales data in a PostgreSQL database
- Provide endpoints to fetch sales data / aggregated data
- Visualize sales data using Chart.js
- Graceful shutdown
- Health check endpoint

## Prereqs

- Go 1.17 or later
- PostgreSQL
- Docker (optional, for containerization)

## Getting Started

### Clone the Repo

```sh
git clone https://github.com/noahharshbarger/sales-data-visualization.git
cd sales-data-visualization
```

### Set up the DB

1. Install PSQL if you haven't already.
2. Create a database and a user:

```
CREATE DATABASE analyticsdb;
CREATE USER analyticsuser WITH ENCRYPTED PASSWORD 'secret';
GRANT ALL PRIVILEGES ON DATABASE analyticsdb TO analyticsuser;
```

### Congiure Environment variables
Create a `.env` file in the root directory and add the following:

```
DB_USER=analyticsuser
DB_PASSWORD=secret
DB_NAME=analyticsdb
DB_SSLMODE=disable
```

### Run the Application

`go run main.go`

### Access the Application

Open the browser and navigate to `http://localhost:8080` to view the sales data visualization.

## API Endpoints

### Get Sales Data

- URL: `/sales`
- Method: `GET`
- Query Parameters:
    - `page` (optional): Page number for pagination
- Response: JSON array of sales data

### Get Aggregated Data

- URL: `/aggregated`
- Method: `GET`
- Response: JSON array of aggregated sales data

### Add Sales Data
- URL: `/add`
- Method: `POST`
- Request Body: JSON object representing a sale (sales_data.csv for ex)

### Health Check
- URL: `/health`
- Method: `GET`
- Response: "OK" if the server is running

## Docker Support

To run the application in a Docker container, follow these steps:
1. Build the Docker image:
    `docker build -t sales-data-visualization .`
2. Run the Docker Container:
    `docker run -p 8080:8080 --env-file .env sales-data-visualization`

## Project Structure
```
.
├── Dockerfile
├── README.md
├── main.go
├── sales_data.csv
└── static
    ├── index.html
    └── styles.css
```

### Contributing

Contributions are welcome! Please open an issue or submit a pull request.

### License

This project is licensed under the MIT License.

