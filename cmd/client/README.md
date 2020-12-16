# GoLogX CLI Tool
Command prompt tool to fetch system logs 

### List of commands:
#### help  
**Usage:** logx help  
Prints out a list of commands you can use.  

#### status  
**Usage:** logx status -postgres postgresql://login:password@127.0.0.1/dbname --services serviceA, serviceB
Will print the status of matched services (last seen, alive, etc.).
It can have 0 - N service names. No service name input will output all services.

#### search  
**Usage:** logx search --postgres postgresql://login:password@127.0.0.1/dbname --service serviceA --dateafter 2010-01-01 00:00:00 -datebefore 2010-01-05 00:00:00 --limit 100 --matcher machinename --orderby columnA, columnB  
Allows to search for logs matching considering the search criteria. Service flag is mandatory. Default limit is 50.

#### details  
**Usage:** logx details --postgres postgresql://login:password@127.0.0.1/dbname --id 12345
Allows log details searching by ID. Id flag is mandatory.