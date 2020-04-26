# diet-orchestrator

This project serves the purpose of orchestrating a food delivery system - by integrating with such, we are able to make the food move from the order to the kitchen up to a drone delivery system! :helicopter:

## The idea

The flow of the project is almost a simple finite state machine. The food can be thought of as having 3 main states (and some substates under those): `order`, `kitchen` and `shipment`. 

To keep track of the current order, we schedule a job that polls it every **x** seconds. On the result, for each order, we check the current status and whether or not it is ready to go to the next stage. Basically:

- When an order is **new**, we will create a **kitchen order**;
- When the **kitchen order** is set to **packaged**, we update the initial order and create a **shipment order**;
- When that **shipment order** points to either **delivered** or **rejected**, we update the initial order and be done with it.

As the food services kept no relations whatsoever, we use firestore to save the relations between an **order** and its **kitchen** and **shipment** orders.

## Assumptions

There were a few assumptions for this example to be developed. There's a small limbo when an order goes from one service to another. For instance, whenever we send a package from the kitchen to the delivery service, it's neither just packaged nor in flight already. For the sake of brevity,  we assume that the `NEW` state in delivery will set the initial order to `IN_FLIGHT`.

Another assumption is that the food service kept no relation between the services. Therefore, we had to store those in another firestore collection.


## Structure

The `mocks` folder contain the necessary gRPC mocks used for testing. The `database` package contains operations regarding firestore.  The `processors` is the most important folder, containing an orchestrator that distributes the work between the kitchen, order and shipment. 

## How to run

First, clone the project and go to the folder. This project was made using go mod, so just check if you have at least go 11.
Before we run the service, let's generate a localhost certificate to use for the communication between the food service and this scheduler.

```
openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout localhost.key -out localhost.pem -subj "/C=US/CN=localhost"
openssl x509 -outform pem -in localhost.pem -out localhost.crt
```

Now you should have your `localhost.crt` and `locahost.key`. Let us now build, get firebase up & export its connection and run the program:

```
go build
gcloud beta emulators firestore start --host-port=0.0.0.0 &
export FIRESTORE_EMULATOR_HOST=0.0.0.0:8080
./diet-scheduler -certificate=../localhost.crt -server=localhost:9000 -interval 45
```
**Note**: for better logs, run the firestore emulator in another terminal.

The flags should be self-explanatory, and that is an extensive list of them. If in doubt, just run `./diet-scheduler -help`.

You can start the food program before or after starting this one, no issues with that. 

## How to run the tests

The tests were made using suites as we needed some work to get the mocks setup. Most of those tests are in the package `processors_test`. In order to run them:

```
cd processors_test
gcloud beta emulators firestore start --host-port=0.0.0.0 &
export FIRESTORE_EMULATOR_HOST=0.0.0.0:8080
go test
```

**Note**: for better logs, run the firestore emulator in another terminal.

You do not need the food service running as we mock it; however, you should need firestore up and running. We perform a cleanup after every round of tests. The author could not find a properly good way to mock firestore. 
