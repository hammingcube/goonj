function showLoggedInState(nickname) {
    document.getElementById('btn-login').textContent = 'Log out';
    document.getElementById('logged-in-box').style.display = 'inline';
    document.getElementById('logged-in-box-anon').style.display = 'none';
    document.getElementById('nick').textContent = nickname;
}

function showStartCodingLink(data) {
  $("#the_link").attr("href", "cui/"+ data.ticket_id);
  $("#the_link").attr("target", "_blank");
  $("#the_link").show();
}


$(document).ready(function() {
    var lock = new Auth0Lock(
        // These properties are set in auth0-variables.js
        AUTH0_CLIENT_ID,
        AUTH0_DOMAIN
    );

    var userProfile;
    var nickname = sessionStorage.getItem('nickname');
    //alert(nickname);
    if(nickname) {
      showLoggedInState(nickname);
    }

    document.getElementById('btn-login').addEventListener('click', function() {
      $("#the_link").attr("href", "")
      $("#the_link").hide()
      if(this.textContent == "Log out") { 
          localStorage.clear();
          sessionStorage.clear();
          document.getElementById('btn-login').textContent = 'Log in';
          document.getElementById('logged-in-box').style.display = 'none';
          document.getElementById('logged-in-box-anon').style.display = 'inline';
          return
      }
      lock.show(function(err, profile, token) {
        if (err) {
          // Error callback
          console.error("Something went wrong: ", err);
          //alert("Something went wrong, check the Console errors");
        } else {
          // Success calback
          // Save the JWT token.
          localStorage.setItem('userToken', token);
          localStorage.setItem('userId', profile.user_id);


          // Save the profile
          userProfile = profile;
          nickname = userProfile.nickname;
          sessionStorage.setItem('nickname', nickname);
          showLoggedInState(nickname);
        }
      });
    });


  document.getElementById('btn-api').addEventListener('click', function() {
    // Just call your API here. The header will be sent
    console.log("Hello real world");
    $.ajax({
      url: "/secured/ping",
      data: {"user_id": localStorage.getItem('userId')},
      beforeSend: function(xhr) {
        xhr.setRequestHeader("Authorization", 'Bearer ' + localStorage.getItem('userToken'));
      },
      error: function(err) {
        // error handler
        console.log(JSON.stringify(err));
      },
      success: function(data) {
        // success handler
        console.log(JSON.stringify(data));
        showStartCodingLink(data);
      }
    });
  });

  document.getElementById('btn-api-anon').addEventListener('click', function() {
    // Just call your API here. The header will be sent
    console.log("Hello anonymous user");
    $.ajax({
      url: "/cui/new",
      //data: {"user_id": localStorage.getItem('userId')},
      beforeSend: function(xhr) {
        //xhr.setRequestHeader("Authorization", 'Bearer ' + localStorage.getItem('userToken'));
      },
      error: function(err) {
        // error handler
        console.log(JSON.stringify(err));
      },
      success: function(data) {
        // success handler
        console.log(JSON.stringify(data));
        showStartCodingLink(data);
      }
    });
  });

});
