# go-rest

1.	Current GitHub repository https://github.com/eshtukin/go-rest is used for Pull Requests monitoring

2.	Twitter account with API access: @goplay25984392

3.  Before running the program you have to define 4 environment variables (which are required for access Twitter dev account via API):

        TWITTER_CONSUMER_KEY
        TWITTER_CONSUMER_SECRET
        TWITTER_ACCESS_TOKEN
        TWITTER_ACCESS_TOKEN_SECRET

4.	Credentials for them are not kept in GitHub, and will be provided separately in twitter.env file using template 

        <env variable>=<value>
  
    The quickest way to define env variable on Linux is:
    
        >for line in $(cat twitter.env)
      
    On Windows you can create batch file with 4 rows like this:
    
		    set <env variable>=<value>
        
5.	For Twitter OAuth authorization 3rd party package github.com/mrjones/oauth is used

6.	For symmetry we could use access token authorization on GitHub too (which would allow 5000 requests per hour instead of default 60),       but it’s not implemented yet

7.	You start the program running command

        >go run main.go
        
    The program gathers PRs (either opened after the previous run, or on the first run - all open PRs).
    
8.	The timestamp of current run is kept in the file baseline.txt for future run

9.	Unit tests are yet to be written

