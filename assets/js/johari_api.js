

var GetWordsUrl = "/get?words=t";
var GetWindowsUrl = "/get?windows=t";
var GetSubmissionStatsUrl = "/get?submissions=t&pane=";
var GetWindowPaneDataUrl = "/get?panedata=t&pane=";
var GetUserInfoUrl = "/get?user=t&pane=";
var GetSubmissionHistory = "/get?history=t&pane=";
var GetPreviousWindows = "/get?previouswindows=t"
var LoginUserURL = "/login/submit"
var LoginRegisterURL = "/login/register"

var CreateWindowUrl = "/post?new=t";
var FeedbackWindowUrl = "/post?feedback=t&pane=";
var SelectedWords = {};
var CurrentPane = "";
var AllPanes = [];
var WindowHistoryData = {};

function toogleLoadSpinner() {
 if ( $('#loadspinner').hasClass("loader") ) {
  $('#loadspinner').removeClass('loader').addClass('unloaded');
 } else {
  $('#loadspinner').removeClass('unloaded').addClass('loader');
 }  

}



function GetAllWords() {
  var target = document.getElementById('GetAllWordsSpinner');
  toogleLoadSpinner();

  $.ajax({
    url: GetWordsUrl,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      //#WordTableContents
      var tableHTML = '<table class="table table-bordered"><tbody>';
      var row = [5];
      for (var i = 0; i < data.words.length; i++) {
        if ( i%5 == 0 && i != 0 || i == data.words.length-1 ) {
          // if this is the last record then make sure to add it
          if (i == data.words.length-1) {
            row[i%5] = '<td id="adj-' + data.words[i] + '" onclick="ToggleWord(\'' + data.words[i] + '\');">' + data.words[i] + '</td>';    
          }
          
          tableHTML += "<tr>" + row.join("\n") + "</tr>";
          for (var x = 0; x < row.length; x++) {
            row[x] = "";
          }
        } 
        row[i%5] = '<td id="adj-' + data.words[i] + '" onclick="ToggleWord(\'' + data.words[i] + '\');">' + data.words[i] + '</td>';  
      }
      tableHTML += "</tbody></table>";
      $('#WordTableContents').html(tableHTML);
    },
    error: function(data) {
      toogleLoadSpinner();
        alert('FAILED to fetch words: ' + data.responseJSON.errmessage);
    }});
}

// if key exists in global SelectedWords{} then toggle it otherwise add key and enable
function ToggleWord(word) {
  selectedColor = "#ccc";
  unselectedColor = "#fff";
  if (SelectedWords.hasOwnProperty(word) && SelectedWords[word]) {
    SelectedWords[word] = false;
    $('#adj-' + word).css("background-color", unselectedColor);
    return
  } 
  SelectedWords[word] = true;
  $('#adj-' + word).css("background-color", selectedColor);
}

/*
  Creates a new window from users word selection
  
  - make sure nickname was given
  - make sure enough words were selected 
  
  sumbit request and response should redicrt user to their new window
*/
function CreateNewWindow() {
  
  if ( $('#Nickname').val() == "" ) {
    alert("You must give a nickname");
    return;
  }
  
  var words = []
  for (var key in SelectedWords) {
    if(SelectedWords.hasOwnProperty(key) && SelectedWords[key] ) {
      words.push(key);
    }
  }
  
  if (words.length < 5 || words.lenght > 20) {
    alert("You must choose 5 to 20 words in order for this thing to work ;-)");
    return;
  }
  
  var PostDATA = {
    nickname: $('#Nickname').val(),
    words: words
  } 

  toogleLoadSpinner();
  
  $.ajax({
    url: CreateWindowUrl,
    type: 'post',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      CurrentPane = data.pane;
      window.location.href = "/window?pane=" + CurrentPane;
    },
    error: function(data) {
      toogleLoadSpinner();
      alert('FAILED to create window: ' + data.responseJSON.errmessage);
    },
    data: JSON.stringify(PostDATA)
  });
}

function SubmitFeedback() {
  var pane = location.search.split('pane=')[1]
  if ( !pane ) {
    alert("Could not find pane ID in url string");
    return;
  }
  
  var words = []
  for (var key in SelectedWords) {
    if(SelectedWords.hasOwnProperty(key) && SelectedWords[key] ) {
      words.push(key);
    }
  }
  
  if (words.length < 5 || words.lenght > 10) {
    alert("You must choose 5 to 10 words in order for this thing to work ;-)");
    return;
  }
  
  var PostDATA = {
    pane: pane,
    words: words
  } 
  
  toogleLoadSpinner();
  
  $.ajax({
    url: FeedbackWindowUrl + pane,
    type: 'post',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      window.location.href = "/thanks";
    },
    error: function(data) {
      toogleLoadSpinner();
      alert(data.responseJSON.errmessage);
    },
    data: JSON.stringify(PostDATA)
  });
  
}

function PopulatePreviousWindows() {
  $.ajax({
    url: GetPreviousWindows,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      var items = ""
      for (var i = 0; i < data.length; i++) {
        items += '<li><a href="/window?pane='+data[i].pane+'"> '+data[i].nickname+'</a></li>';
      }
      if (items != "") {
        $('#PreviousWindowItems').html(items);
      }
    },
    error: function(data) {
      alert(data.responseJSON.errmessage);
    }
  });
}

function GetUserWindows() {
  
  toogleLoadSpinner();

  $.ajax({
    url: GetWindowsUrl,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      if ( data.length >= 1 ) {
        CurrentPane = data[0].session;
      } else {
        // no sessions so inform users 
        DisplayEmptyPanel();
        return
      }
      setCurrentPain(); // if pane is set then change to that pane
      AllPanes = data;
      var sharableLink = BASEURL+'/feedback?pane='+CurrentPane;
      $('#shareable-link').html('<ul><li><a href="' + sharableLink + '" target="_blank">'+ sharableLink + '</a>');
      DisplayCurrentPane();
    },
    error: function(data) {
      toogleLoadSpinner();
      alert('FAILED to create window: ' + data.responseJSON.errmessage);
    }
  });
}

function setCurrentPain() {
  var query = location.search.substr(1);
  var result = {};
  query.split("&").forEach(function(part) {
    var item = part.split("=");
    result[item[0]] = decodeURIComponent(item[1]);
  });
  if ( result.hasOwnProperty("pane") ) {
    console.log("updated pane with " + result["pane"] );
    CurrentPane = result["pane"];
  }
  console.log("pane is default with " + CurrentPane );
}

function DisplayCurrentPane() {
  DisplayPane(CurrentPane);
}

function DisplayPane(pane) {
  BuildJohariWindow();
  BuildCliftonWindow();
  GetSumissionStats(pane);
  PopulateJCWindows(pane);
  DisplayHistory(pane);
}

function GetSumissionStats(pane) {
  
  toogleLoadSpinner();
  
  $.ajax({
    url: GetSubmissionStatsUrl + pane,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      $('#SumbissionStats').html('<table class="table table-sm table-inverse"><thead><tr><th>Number Of User Submissions</th><th>' + parseInt(data.submissions) + '</th></tr></thead></table>' );
    },
    error: function(data) {
      toogleLoadSpinner();
      alert(data.responseJSON.errmessage);
    }
  });
}

function BuildJohariWindow() {
  //JohariWindow
  $('#JohariWindow').html('<table class="table table-bordered">' +
    '<tr>' +
      '<td class="tabletitle">Johari Window</td>' +
      '<td class="windowtitle">Known to Self</td>' +
      '<td class="windowtitle">Not Known to Self</td>' +
    '</tr>' +
    '<tr>' +
      '<td class="windowtitle">Known to Others</td>' +
      '<td class="windowwords"><div id="JohariWindow-arena"></div></td>' +
      '<td class="windowwords"><div id="JohariWindow-blind"></div></td>' +
    '</tr>' +
    '<tr>' +
      '<td class="windowtitle">Not Known to Others</td>' +
      '<td class="windowwords"><div id="JohariWindow-facade"></div></td>' +
      '<td class="windowwords"><div id="JohariWindow-unknown"></div></td>' +
    '</tr>' +
    '</table>');
}
function BuildCliftonWindow() {
  $('#CliftonWindow').html('<table class="table table-bordered">' +
    '<tr>' +
      '<td class="tabletitle">Clifton Themes</td>' +
      '<td class="windowtitle">Known to Self</td>' +
      '<td class="windowtitle">Not Known to Self</td>' +
    '</tr>' +
    '<tr>' +
      '<td class="windowtitle">Known to Others</td>' +
      '<td class="windowwords"><div id="CliftonWindow-arena"></div></td>' +
      '<td class="windowwords"><div id="CliftonWindow-blind"></div></td>' +
    '</tr>' +
    '<tr>' +
      '<td class="windowtitle">Not Known to Others</td>' +
      '<td class="windowwords"><div id="CliftonWindow-facade"></div></td>' +
      '<td class="windowwords"><div id="CliftonWindow-unknown"></div></td>' +
    '</tr>' +
    '</table>');
}

function PopulateJCWindows(pane){
  toogleLoadSpinner();
  
  $.ajax({
    url: GetWindowPaneDataUrl + pane,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      $('#JohariWindow-arena').html(data.johari.arena.join(", "))
      $('#JohariWindow-blind').html(data.johari.blind.join(", "))
      $('#JohariWindow-facade').html(data.johari.facade.join(", "))
      $('#JohariWindow-unknown').html(data.johari.unknown.join(", "))
      
      $('#CliftonWindow-arena').html(data.clifton.arena.join(", "))
      $('#CliftonWindow-blind').html(data.clifton.blind.join(", "))
      $('#CliftonWindow-facade').html(data.clifton.facade.join(", "))
      $('#CliftonWindow-unknown').html(data.clifton.unknown.join(", "))
    },
    error: function(data) {
      toogleLoadSpinner();
      alert(data.responseJSON.errmessage);
    }
  });
}

function DisplayHistory(pane) {
  toogleLoadSpinner();
  
  $.ajax({
    url: GetSubmissionHistory + pane,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      WindowHistoryData = data.users;
      FilterHistory(false);
      
    },
    error: function(data) {
      toogleLoadSpinner();
      alert(data.responseJSON.errmessage);
    }
  });
}

function FilterHistory(doFilter) {
  var historyTableHTML = '<table class="table table-bordered">' +
    '<thead><tr><th>User</th><th>Themes</th><th>Words</th></tr></thead>';
  
  var userFilter = $('#SearchUser').val();
  var themeFilter = $('#SearchTheme').val();
  var wordsFilter = $('#SearchADJ').val();
    
  for (var key in WindowHistoryData) {
    if(WindowHistoryData.hasOwnProperty(key)) {
      if (doFilter) {
        if (userFilter != "" && userFilter != key) {
          continue;
        }
        if (themeFilter != "" && !FilterCheck(themeFilter, WindowHistoryData[key].themes)) {
          continue;
        }
        if (wordsFilter != "" && !FilterCheck(wordsFilter, WindowHistoryData[key].words)) {
          continue
        }
      }
      historyTableHTML += '<tr>' +
      '<td>' + key + '</td>' + 
      '<td>' + WindowHistoryData[key].themes.join(", ") + '</td>' + 
      '<td>' + WindowHistoryData[key].words.join(", ") + '</td>' + 
      '</tr>';
    }
  }
  historyTableHTML += "</table>";
  $('#SumbissionHistory').html(historyTableHTML);
}

function FilterCheck(filter, s) {
  for (var i = 0; i < s.length; i ++) {
    if (s[i] == filter) {
      return true;
    }
  }
  return false;
}

function ClearFilterHistory() {
  $('#SearchUser').val("");
  $('#SearchTheme').val("");
  $('#SearchADJ').val("");
  FilterHistory(false);
}

/*function DisplayPanel() {
  var sharableLink = BASEURL+'/feedback?pane='+CurrentPane;
  $('#DisplayPane').html('<div class="panel-group">' +
    '<div class="panel panel-success">' +
      '<div class="panel-heading"><center>Welcome to your very own personality window pane!</center></div>' +
      '<div class="panel-body">' +
        '<p><center><img src="/img/johari-window.PNG" width="300" ALIGN=center></center></p>' +
        '<p><ul>' + 
          '<li><b>Arena:</b> Adjectives/Themes that are selected by both you and your peers are placed into the Open or Arena quadrant</li>' + 
          '<li><b>Facade:</b> Adjectives/Themes selected only by you and not by any of your peers are placed into the Hidden or Façade quadrant</li>' +
          '<li><b>Blind:</b> Adjectives/Themes that are not selected by you but selected by your peers are placed into the Blind Spot quadrant</li>' +
          '<li><b>Unknown:</b> Adjectives/Themes that were not selected by either you or your peers remain in the Unknown quadrant</li>' +
        '</ul></p>' +
        '<p>Please use this link to request feedback from your peers</p>' +
        '<p><ul><li><a href="'+sharableLink+'">'+sharableLink+'</a></li></ul></p>' +
      '</div>' +
    '</div>' +
  '</div>');
}*/

function DisplayEmptyPanel() {
  $('#DisplayPane').html('<div class="panel-group">' +
    '<div class="panel panel-warning">' +
      '<div class="panel-heading"><center>You currently have no sessions to display</center></div>' +
      '<div class="panel-body">' +
        '<p>Please return to the <a href="/">main site</a> and start a new session </p>' +
      '</div>' +
    '</div>' +
  '</div>');
}

function DisplayEmptyFeedbackPanel() {
  $('#DisplayPane').html('<div class="panel-group">' +
    '<div class="panel panel-warning">' +
      '<div class="panel-heading"><center>You might be in the wrong place</center></div>' +
      '<div class="panel-body">' +
        '<p>Please return to the <a href="/">main site</a> or check with your peer regarding the quality of the url they sent you</p>' +
      '</div>' +
    '</div>' +
  '</div>');
}

function DisplayFeedbackItems() {
  var pane = location.search.split('pane=')[1]
  if ( !pane ) {
    alert("There was no 'pane=<sessionid>' in the url string please make sure you have the correct url")
    DisplayEmptyFeedbackPanel();
    $('#feedbackform').html("");
    return
  }
  DisplayUserProfile(pane);
}

function DisplayUserProfile(pane){
  toogleLoadSpinner();
  
  $.ajax({
    url: GetUserInfoUrl + pane,
    type: 'get',
    dataType: 'json',
    success: function (data) {
      toogleLoadSpinner();
      var emailname = '<span style="color:blue">' + data.email + "</span>"
      $('#UserProfileSubmissionHelp').html('User with email address of ' + emailname+ ' has asked you to kindly submit personality feedback' +
      '<p><ol>' +
        '<li>Please select between 5 to 20 words that you think best describes ' + emailname + ' personality</li>' +
        '<li>Click submit at the bottom once you have finished making your selection</li>' +
      '</ol></p>' +
      '<p>Please note this is not an anonymous submission and your feedback is most welcome and appreciated!  The primary goal of this exercise is to help give the subjects insightful external perspectives that enable them to grow as individuals</p>');
    },
    error: function(data) {
      toogleLoadSpinner();
      alert(data.responseJSON.errmessage);
    }
  });
}


function registerUser() {

  var registerRequest = {
    "user": $('#usernameField').val(),
    "password": btoa($('#userpasswordField').val())
  };

  $.ajax({
    url: LoginRegisterURL,
    type: 'POST',
    dataType: 'json',
    success: function (data) {
      window.location.href = "/login";
    },
    error: function(data) {
      alert(data.responseJSON.errmessage);
    },
    data: JSON.stringify(registerRequest)
  });
}

function loginUser() {

  var loginRequest = {
    "user": $('#usernameField').val(),
    "password": btoa($('#userpasswordField').val())
  };

  $.ajax({
    url: LoginUserURL,
    type: 'POST',
    dataType: 'json',
    success: function (data) {
      window.location.href = "/";
    },
    error: function(data) {
      alert(data.responseJSON.errmessage);
    },
    data: JSON.stringify(loginRequest)
  });
}