# MVC Assignment

Run with `docker compose up`

Open "http://localhost:3000/swagger/index.html"

If not using docker:
1. Configure environment variables in `.env` file (see `.env.sample` for reference)
2. Generate swagger docs with `make swagger` (must have swag cli installed) (optional)
3. Run the application with `make run`

A default admin user is created with the following credentials:
- Username: `admin`
- Password: `g`

Benchmarking:
```
# ensure you are in the root repo directory
./testbench/bench.sh
```
