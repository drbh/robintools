# capgains

sometimes your like "wait... what are my capital gains? how much will I own in taxes?"  

This helps get part of the way there!  


This is what the output for a single ticker looks like
```
##########################
Plug Power, Inc. Common Stock
PLUG
-
endquantity: 60
executions:
- price: 8.75
  quantity: 60
  date: 2020-07-15T15:49:13.355Z
totalgainloss: -525
averageprice: 8.75
quantity: 60
gainloss: -525
difference: 0
side: buy
state: filled
time: 2020-07-15T15:49:13.355Z
-
endquantity: 0
executions: []
totalgainloss: 38.69999999999993
averageprice: 9.395
quantity: 60
gainloss: 563.6999999999999
difference: 0.9523501041666668
side: sell
state: filled
time: 2020-07-16T14:40:36.404Z
-
endquantity: 5
executions:
- price: 31.94
  quantity: 5
  date: 2020-12-29T18:15:11.193Z
totalgainloss: -121.00000000000009
averageprice: 31.94
quantity: 5
gainloss: -159.70000000000002
difference: 0.9523501041666668
side: buy
state: filled
time: 2020-12-29T18:15:11.193Z
-
endquantity: 0
executions: []
totalgainloss: 145.59999999999994
averageprice: 53.32
quantity: 5
gainloss: 266.6
difference: 10.185295706018518
side: sell
state: filled
time: 2021-01-08T22:42:00.742Z
```

You can see the buy and sell, and the cost and gains related to those orders. Using this its easy to sum up all of your capital gains and losses for a period and determine your taxes  


### Usage

```
cd capgains
go run src/main.go
```

you may want to save the output to read through each companys. try adding `> record.txt` to the above command to save the output.

also note, sometimes companies change their tickers and their unique identifer from robinhood. This can cause strange looking captial gains for those companies because this script does not know they are the same company. 