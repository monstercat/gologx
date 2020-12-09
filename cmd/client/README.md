# Go LogX Client
Command prompt tool to fetch system logs 

### List of commands:
#### help  
**Usage:** logx help  
Prints out a list of commands you can use.  

#### status  
**Usage:** logx status [service names...]  
Will print the status of matched services (last seen, alive, etc.).  
It can have 0 - N service names. No service name input will output all services.  

#### search  
**Usage:** logx search [service name] [text match] [date (before/after)]  
Allows to search for logs matching any the the criteria  
