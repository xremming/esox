async function handler(event) {
    const request = event.request;
    const headers = request.headers;
    const host = request.headers.host.value;

    headers["x-original-host"] = { value: host };

    return request;
}
