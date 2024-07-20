# Use the official PostgreSQL image for version 16 on Alpine Linux
FROM postgres:16-alpine

# Optionally, you can add additional configuration or customization here
# For example, you can copy initialization scripts or configuration files
# from your host machine to the container.
ENV POSTGRES_DB=deu
ENV POSTGRES_USERNAME=deu_admin
ENV POSTGRES_PASSWORD=nolodire

# Copy initialization scripts into the container
COPY init-db.sh /docker-entrypoint-initdb.d/

EXPOSE 5432