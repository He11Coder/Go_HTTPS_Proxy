# Go_HTTPS_Proxy
Simple Go HTTP(S) proxy server with its own WEB API and request storage. It can be used to analyze and scan WEB-pages, resources and requests.
This is the ```'WEB Application Security' VK Education Homework```.


## Use the following ```make``` commands to configure and prepare the application:
```bash
make create-cert #to create CA certificate and key
make add-cert-as-trusted #to mark this certificate as trusted in your Linux OS
make docker-build #to build two Docker images: for https-proxy server and its scanning API
```


## Running the application
```bash
make docker-compose-up #this command includes the 'make docker-build' one. Use it to build all the necessary images and run the application in Docker containers
```
After you execute the command above, https-proxy server will listen at ```127.0.0.1:8080``` or ```0.0.0.0:8080``` and WEB API can be accessed at ```127.0.0.1:8000``` or ```0.0.0.0:8000```.

*Note: PostgreSQL is at the port ```8055``` of your ```localhost```. Check ```docker-compose.yaml``` to get the credentials to access (if you need so).*


## Available API
By default, the application is able to serve the following APIs:
* ```GET /requests``` - to get all the requests stored in the database (as table)
* ```GET /requests/{requestID}``` - to get a particular request with the given ID
* ```GET /repeat/{requestID}``` - to repeat a saved request
* ```GET /scan/{requestID}``` - to scan (to dirbust) the request's (defined by ID) host for possible vulnerabilities
