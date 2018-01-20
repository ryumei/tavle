Vue.config.devtools = true
Vue.use(VueSessionStorage)

// Initialize WebSocket connection
var connectWs = function(vueBase) {
    var self = vueBase;

    protocol = (this.location.protocol == 'https:') ? 'wss:' : 'ws:'; 
    vueBase.ws = new WebSocket(protocol + '//' + window.location.host + '/ws/' + vueBase.room);
    
    vueBase.ws.addEventListener('message', function(e) {
        console.log('[DEBUG] Receive data from the server: ' + e.data);
        
        //NOTE: Received data have been sanitized on the serverside
        var msg = JSON.parse(e.data);
        var message = emojione.toImage(msg.message.replace(/\r?\n/g, '<br/>'));

        self.talkTimeline.push({
            avatarImg: (msg.email != "") ? '<img src="https://s.gravatar.com/avatar/' + CryptoJS.MD5(msg.email) + '" />' : '',
            username: msg.username,
            message: message
        });

        var element = document.getElementById('chat-messages');
        element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
    });
}            

var escapeNewline = function(str) {
    return str.replace(/\n/g, "<br/>");
}

// timeline component
Vue.component('timeline', {
    props: ['msg'],
    template: '<div class="post">' + 
        '<div class="chip"><span v-html="msg.avatarImg"></span> {{msg.username}}</div> ' +
        '<span v-html="msg.message"></span></div>',
})

new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        talkTimeline: [], // A running list of chat messages displayed on the screen
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        room: null, // Unique room name
        joined: false // True if email or username have been filled in
    },
    components: {
    },
    created: function() {
        if (this.$session.get("created")) {
            try {
                stub = JSON.parse(this.$session.get("stub"));
                this.email = stub['email'];
                this.username = stub['username'];
                this.room = stub['room'];
                this.joined = true;
                connectWs(this);
                
                console.log("[DEBUG] Reconnected " + this.username + "@" + this.room);
            }
            catch(e) {
                console.log("[ERROR] " + e);
            }
        } else {
            console.log("[DEBUG] not created. Initializing");
            this.$session.set("created", true);            
        }
    },
    methods: {
        send: function (event) {
            if (event.shiftKey) {
                return;
            }
            if (this.newMsg != '') {
                msg = this.newMsg
                this.ws.send(
                    JSON.stringify({
                        email: this.email,
                        username: this.username,
                        room: this.room,
                        message: $('<p>').html(msg).html()  // message should be sanitized on the server side
                    }
                ));
                this.newMsg = ''; // Reset newMsg
                $('#message').val("");  // Force reset Firefox with enter key
            }
        },
        join: function () {
            if (!this.username) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            if (!this.room) {
                this.room = "foyer";
                console.log("[WARN] Use default roomname '" + this.room + "' instead of empty.");
            }
            this.email = $('<p>').html(this.email).text();
            this.username = $('<p>').html(this.username).text();
            this.room = $('<p>').html(this.room).text();
            this.joined = true;
            
            this.$session.set("stub", JSON.stringify({
                username: this.username,
                email: this.email,
                room: this.room
            }));
            this.$session.set("joined", true);

            // Initialize WebSocket connection
            connectWs(this);
        }
    }
});
