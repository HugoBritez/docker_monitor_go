# Usar una imagen base de Go para compilar
FROM golang:1.21-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache gcc musl-dev

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o docker_monitor .

# Imagen final
FROM alpine:latest

# Instalar certificados CA para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copiar el binario compilado
COPY --from=builder /app/docker_monitor /usr/local/bin/

# Crear directorio para logs
RUN mkdir -p /var/log/docker-monitor

# Establecer el comando por defecto
CMD ["docker_monitor"] 