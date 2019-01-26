Team members: Joan Massana, Roberto Carlini

How to run:

- Exercise1

In order to run the program listening in port 6001 and dialing to port 6002, just run:

go run Lab1/exercise1/exercise1.go 6001 6002

- Exercise3

In order to run the program listening in port 6001 and dialing to port 6002, just run:

go run Lab1/exercise2/exercise2.go 6001 6002

The program will assume that the files to be sent are under Lab1/exercise2/files/ and the received files will be saved under Lab1/exercise2/files/received.

- Exercise3

The program assumes that is being executed at the base directory, it is, should be run like this:

go run Lab1/exercise3/exercise3.go configNode1.add

As the exercise 2, the program will assume that the files to be sent are under Lab1/exercise2/files/ and the received files will be saved under Lab1/exercise2/files/received.
Also, it assumes that the configuration files will be under Lab1/exercise3/files (that's why no path is supplied for configNode1.addd)
We plan to fix those assumptions in the next days...

jorge.lobo@upf.edu
lab1_roberto_carlini.zip