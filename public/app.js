Vue.config.devtools = true
Vue.use(VueSessionStorage)

Vue.filter('formatDatetime', function(value) {
    if (value) {       
        return value
    }
})


// Initialize WebSocket connection
var connectWs = function(vueBase) {
    var self = vueBase;

    protocol = (this.location.protocol == 'https:') ? 'wss:' : 'ws:'; 
    vueBase.ws = new WebSocket(protocol + '//' + window.location.host + '/ws/' + vueBase.room);
    
    vueBase.ws.addEventListener('message', function(e) {
        console.log('[DEBUG] Receive data from the server: ' + e.data);
        
        //NOTE: Received data have been sanitized on the server side
        var msg = JSON.parse(e.data);
        if (msg.type == 'members') {
            self.memberCount = msg.count;
            return;
        }
        if (msg.type == 'delete') {
            self.talkTimeline = self.talkTimeline.filter(function(post) {
                return post.id != msg.id;
            });
            self.forgetDeleteToken(msg.id);
            return;
        }
        if (msg.deleteToken) {
            self.saveDeleteToken(msg.id, msg.deleteToken);
        }
        var message = emojione.toImage(msg.message.replace(/\r?\n/g, '<br/>'));

        /*
         * TODO: Switching option reloading or updating
         * e.data で array を受け付ける。
         * List なら、reload として前に付ける。
         * Item なら、new post として後に付ける。
         */

        // Auto scroll only when the user is near the bottom or the message is own post
        var element = document.getElementById('chat-messages');
        var autoScroll = element == null ||
            element.scrollHeight - element.scrollTop - element.clientHeight < 50 ||
            msg.username == self.username;

        self.talkTimeline.push({
            avatarImg: (msg.email != "") ? '<img src="https://s.gravatar.com/avatar/' + CryptoJS.MD5(msg.email) + '" />' : '',
            username: msg.username,
            message: message,
            timestamp: msg.timestamp,
            id: msg.id,
            cancelable: msg.id != null && self.deleteTokens[msg.id] != null
        });

        //TODO: notify new messages have arrived
        //if ($('.toast').length == 0) {
        //Materialize.toast('Unread messages', 3000);
        //}
        if (autoScroll) {
            self.$nextTick(function() {
                var element = document.getElementById('chat-messages');
                if (element) {
                    element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
                }
            });
        }
    });
}            

var escapeNewline = function(str) {
    return str.replace(/\n/g, "<br/>");
}

// timeline component
Vue.component('timeline', {
    props: ['msg'],
    template: '<div class="post">' + 
        '<div class="chip"><span v-html="msg.avatarImg"></span> {{msg.username}} <span class="timestamp">({{ msg.timestamp | formatDatetime }})</span></div> ' +
        '<a v-if="msg.cancelable" class="cancel" href="#" title="Delete" @click.prevent="$emit(\'cancel\', msg.id)"><i class="material-icons tiny">delete</i></a> ' +
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
        joined: false, // True if email or username have been filled in
        memberCount: null, // Number of connections in the room
        deleteTokens: {} // Tokens to delete own messages, keyed by message id
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
                this.deleteTokens = JSON.parse(this.$session.get("deleteTokens") || "{}");
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
        saveDeleteToken: function(id, token) {
            this.deleteTokens[id] = token;
            this.$session.set("deleteTokens", JSON.stringify(this.deleteTokens));
        },
        forgetDeleteToken: function(id) {
            delete this.deleteTokens[id];
            this.$session.set("deleteTokens", JSON.stringify(this.deleteTokens));
        },
        cancel: function(id) {
            if (!confirm('Delete this message?')) {
                return;
            }
            this.ws.send(JSON.stringify({
                type: 'delete',
                id: id,
                deleteToken: this.deleteTokens[id]
            }));
        },
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
