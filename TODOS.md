# Todos

- [ ] Feat: Message from text area creates chat message
    on enter keypressed event catch value of user input textarea
    process event to be of type http.Request
    post it as a chat message via WS. It should be visible in your twitch chat
    optinally, handle response from WS and remove it from the TUI chat.

- [ ] Fix: there is a bug when token is saved to file, but expired already. Current implementation cant handle this scenario. It leaves open connection unused and eventually program crashes
 