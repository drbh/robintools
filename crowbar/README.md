# crowbar

Sometimes you're bull on a company and want to buy some shares or calls.  

You might ask yourself the following questions:

"whats the best way for me to buy exposure to this stock/underlying?"  
"which of these options provide the most leverage?"  
"at what point are the contracts worth more than equities?"  

Basiclly you want your dollars to work as hard as they can and be as leveraged as possible and minimize the distance between the current price and your breakeven.

example of columns an values    
```
SELECTION 				|	0 
EXPIRATION 				|	2021-01-15 
STRIKE 					|	10.00 
PRICE 					|	4.40 
CONTRACTS 				|	22 
SHARES 					|	699 
BREAKEVEN 				|	14.40 
CONTRACTBREAK 				|	14.49 
DISTANCE 				|	0.19 
RELATIVEDISTANCE 			|	0.01 
LEVERAGE 				|	3.15 
LEVERAGETODISTANCE 			|	231.74 
```


### Usage

```
cd crowbar
go run src/main.go
```

you may want to save the output to read through each companys. try adding `> record.txt` to the above command to save the output.

also note, sometimes companies change their tickers and their unique identifer from robinhood. This can cause strange looking captial gains for those companies because this script does not know they are the same company. 