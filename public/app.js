new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        chatContent: '', // A running list of chat messages displayed on the screen
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        room: null, // Unique room name
        joined: false, // True if email and username have been filled in
        uuid: uuidv1()
    },
    created: function() {
        console.log("[DEBUG] created " + this.uuid);
    },
    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        email: this.email,
                        username: this.username,
                        room: this.room,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                this.newMsg = ''; // Reset newMsg
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

            // Initialize WebSocket connection
            var self = this;
            this.ws = new WebSocket('ws://' + window.location.host + '/ws/' + this.room);
            this.ws.addEventListener('message', function(e) {
                console.log('[DEBUG] Receive data from the server: ' + e.data);
                var msg = JSON.parse(e.data);

                avatarImg = (msg.email != "")
                    ? '<img src="' + self.gravatarURL(msg.email) + '">'
                    : '<img src="data:image/svg+xml;base64,' + self.identicon(msg.username) + '">';
                    
                self.chatContent += '<div class="chip">'
                    + avatarImg + msg.username
                    + '</div>'
                    + emojione.toImage(msg.message) + '<br/>'; // Parse emojis
    
                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            });
        },
        gravatarURL: function(email) {
            return 'https://s.gravatar.com/avatar/' + CryptoJS.MD5(email);
        },
        identicon: function(source) {
            secret = this.uuid;
            hash = CryptoJS.enc.Base64.stringify(CryptoJS.HmacSHA1(source, secret));
            return new Identicon(hash, {format: 'svg'}).toString();
        }
    }
});