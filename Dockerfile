FROM golang:1.27.2 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY annotations ./annotations
COPY metrics ./metrics
COPY measurement ./measurement
COPY lifecycle ./lifecycle
COPY prebuilt ./prebuilt
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -o /out/goo ./cmd/goo

FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
COPY --from=go-build /out/goo /usr/local/bin/goo
ENV PYTHONUNBUFFERED=1
CMD ["python", "server.py"]
