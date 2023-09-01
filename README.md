# Kolla Project 1

### Description 
<p>
This project was designed as a learning experience using the KollaConnect platform. The idea was to integrate two softwares into a usable workflow to improve the flow between the two and minimize human intervention. 
</p>
---

### Execution 
For this project, the two softwares that were chosen were Monday.com and BambooHR. The idea was to take the time-off requests page from BambooHR and integrate it with Monday.com to provide an organized, visually-appealing way to view the time-off requests for a company. 
<br>
Requests were made to Kolla to get both the Monday.com and BambooHR credentials to get each API Key that would be needed to make requests to each account. 
<br>
Requests were then made to Monday.com to get the correct board. Then, requests were made to BambooHR to get employee time-off requests (in a specific date range) and then those requests were were added one by one into the Monday.com board. 
<br>
To allow for updating, when a time-off request was added to the Monday.com board, the item ID was saved and stored in a text file. Then, for each update, the old item IDs were read in and used to delete the old items from the Monday.com board and new item IDs were then generated and added to a text file when the requests were added again. 
<br>
Error checking and overall code flow was done in the way that I saw it most useful. 
--- 

### References and Links
<p>
Kolla: <a href="https://getkolla.com">getkolla.com</a>
<br>
Kolla documentation: <a href="https://docs.getkolla.com/kolla/">docs.getkolla.com</a>
<br><br>
Monday.com: <a href="https://monday.com">monday.com</a>
<br>
Monday.com API reference: <a href="https://developer.monday.com/api-reference/docs">developer.monday.com/api-reference/docs</a>
<br><br>
BambooHR: <a href="https://bamboohr.com">bamboohr.com</a>
<br>
BambooHR API reference: <a href="https://documentation.bamboohr.com/reference/">documentation.bamboohr.com/reference</a>
</p>
---
