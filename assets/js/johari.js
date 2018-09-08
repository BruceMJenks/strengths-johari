

function checkURLParms() {
        var params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});

        if ( params['state'] ) {

                // if base64 generates an = char then it gets converted to html code %3D.  Need to fix that before decoding the string
                encodedstring = params['state'].replace('%3D', '=');
                var decodedstring =  atob(encodedstring);

                var kvs = decodedstring.split("?");
                var pmap = {};
                for ( i = 0; i < kvs.length; i++) {
                        var kv = kvs[i].split("=");
                        pmap[kv[0]] = kv[1];
                }
                params = pmap;
        }

        if ( params['feedbackpane'] ) {
                console.log("setting location to /feedback?pane=" + params['feedbackpane']);
                window.location.href = "/feedback?pane=" + params['feedbackpane'];
        }
}