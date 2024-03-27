// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

function NewTemplates(parentEl, templatesOpts) {
  var fileObjects = [];
  var configBoxesCount = 0;

  function resetFiles() {
    fileObjects = [];
    configBoxesCount = 0;
    $(".config-boxes", parentEl).empty();
  }

  function addFile(e, opts) { // opts = {text, name, focus}
    var configId = "config-box-"+configBoxesCount;

    if (!opts.name) {
      opts.name = "config-" + (configBoxesCount+1) + ".yml";
    }

    if (!opts.text) { opts.text = ""; }

    $(".config-boxes", parentEl).append(
      '<div id="'+configId+'-box">'+
        '<button id="'+configId+'-delete" class="button">x</button>'+
        '<input id="'+configId+'-name" value="'+opts.name+'"/>'+
        '<textarea id="'+configId+'-data">'+opts.text+'</textarea><br/>'+
      '</div>'
    );

    var editor = CodeMirror.fromTextArea(document.getElementById(configId+"-data"), {
      lineNumbers: true,
      viewportMargin: Infinity
    });

    editor.setOption("extraKeys", {
      Tab: function(cm) { cm.replaceSelection("  "); },
      "Shift-Enter": function(cm) { evaluate(); },
    });

    var file = {
      id: configId,
      name: document.getElementById(configId+"-name"),
      data: editor
    };

    file.data.on('change', function(){
      evaluate();
    });

    $("#"+configId+"-name").on("change paste keyup", function() {
      evaluate();
    });

    $("#"+configId+"-delete").click(function() {
      for (var i in fileObjects) {
        if (fileObjects[i].id == configId) {
          $("#"+configId+"-box").remove();
          fileObjects.splice(i, 1);
          evaluate();
          return false;
        }
      }
      return false;
    });

    fileObjects.push(file);
    configBoxesCount++

    if (opts.focus) {
      file.data.focus();
    }
  }

  function setFiles(files) {
    files.sort(function(a,b) {
      var aIsReadme = a.name.startsWith("README.");
      var bIsReadme = b.name.startsWith("README.");
      // 1 indicates b precedes a
      if (aIsReadme && bIsReadme) { return a.name > b.name ? 1 : -1 }
      if (aIsReadme && !bIsReadme) { return -1 }
      if (!aIsReadme && bIsReadme) { return 1 }
      if (!aIsReadme && !bIsReadme) { return a.name > b.name ? 1 : -1 }
      throw "unreachable";
    })

    resetFiles();
    for (var j in files) {
      addFile(null, files[j])
    }
    evaluate();
  }

  var latestReqId = 0;
  var lastAppliedReqId = -1;

  $(".config-form", parentEl).submit(function(ev) {
    var files = [];
    for (i in fileObjects) {
      files.push({
        name: fileObjects[i].name.value,
        data: fileObjects[i].data.getValue(),
      })
    }

    var currReqId = latestReqId++;

    if (templatesOpts.preEvaluateCallback) {
      templatesOpts.preEvaluateCallback(currReqId);
    }

    $.ajax({
      type: "POST",
      url: "/template",
      contentType:"application/json; charset=utf-8",
      dataType: "json",
      data: JSON.stringify({files: files}),

      success: function(data) {
        if (currReqId <= lastAppliedReqId) {
          return
        }
        lastAppliedReqId = currReqId;

        if (data.errors) {
          $(".output", parentEl).text(data.errors);
          $(".output", parentEl).wrapInner('<span class="output-errors"/>');
        } else {
          var result = "";
          for (var i in data.files) {
            result += "<div><h3>"+data.files[i].name+"</h3>" +
              "<textarea>"+data.files[i].data+"</textarea></div>";
          }
          $(".output", parentEl).html(result);
          $('.output textarea', parentEl).each(function() {
            CodeMirror.fromTextArea(this, {
              lineNumbers: true,
              viewportMargin: Infinity,
              readOnly: true
            })
          });
        }

        if (templatesOpts.postEvaluateCallback) {
          templatesOpts.postEvaluateCallback(currReqId);
        }
      },

      error: function(jqxhr, textStatus) {
        if (currReqId <= lastAppliedReqId) {
          return
        }
        lastAppliedReqId = currReqId;

        // Seems like sometimes textStatus is not very descriptive...
        if (textStatus == "error" && jqxhr.status == 0) {
          textStatus = "Unable to establish connection to the server"
        }
        $(".output", parentEl).text("An error occured: " + textStatus);
        $(".output", parentEl).wrapInner('<span class="output-errors"/>');

        if (templatesOpts.postEvaluateCallback) {
          templatesOpts.postEvaluateCallback(currReqId);
        }
      },
    });

    return false;
  });

  function evaluate() {
    $(".config-form", parentEl).submit();
  }

  return {
    resetFiles: resetFiles,
    addFile: addFile,
    evaluate: evaluate,
    setFiles: setFiles,
  }
}

function NewExamples(parentEl, templates, exampleLocation, blocker) {
  function load(id, opts){
    blocker.on();

    $.get('/examples/' + id, function(data) {
      var content = JSON.parse(data);

      if (opts.preDoneCallback) opts.preDoneCallback(id);

      var files = [];
      for (var j in content.files) {
        files.push({
          name: content.files[j].name,
          text: content.files[j].content
        });
      }
      templates.setFiles(files);

      blocker.off();
      if (opts.scrollIntoView) parentEl[0].scrollIntoView();
    });
  }

  $.get("/examples", function(data) {
    var exampleSets = JSON.parse(data);

    exampleSets.forEach(exampleSet => {

      $(".button-container").append(
          '<button type="button" class="button example-set-button" name="' + exampleSet.id + '">' +
           exampleSet.display_name + '</button>'
      );

      $('button[name="' + exampleSet.id + '"]', parentEl).click(function () {
        $('.example-set#example-set-' + exampleSet.id).toggleClass("expanded");
        $('.example-set-button[name="' + exampleSet.id + '"]').toggleClass("expanded");
        return false;
      });

      $(".example-sets").append(
          '<div class="example-set" id="example-set-' + exampleSet.id + '">' +
            '<h3 class="example-set-name">' + exampleSet.display_name + '</h3>' +
            '<p class="example-set-description">' + exampleSet.description + '</p>' +
            '<ol class="dropdown-content" id="' + exampleSet.id + '"></ol>' +
          '</div>'
      );

      var examples = exampleSet.examples
      for (var i = 0; i < examples.length; i++) {
        $("ol#" + exampleSet.id, parentEl).append(
            '<li><a class="item" href="#" data-example-id="' +
            examples[i].id + '">' + examples[i].display_name + '</a></li>');
      }

      $('.dropdown-content .item', parentEl).click(function (e) {
        var example = $(this).data("example-id");
        load(example, {scrollIntoView: true});
        exampleLocation.set(example);
        return false;
      });
    })
    
    $('button[name="' + exampleSets[0].id + '"]', parentEl).click();
  });

  return {
    load: load,
  }
}

function NewGist(parentEl, templates, gistLocation, blocker) {
  function load(id, opts){
    blocker.on();

    $.get('https://api.github.com/gists/' + id, function(data) {
      if (opts.preDoneCallback) opts.preDoneCallback(id);

      var files = [];
      for (var j in data.files) {
        files.push({
          name: data.files[j].filename,
          text: data.files[j].content
        })
      }
      templates.setFiles(files);

      blocker.off();
      if (opts.scrollIntoView) parentEl[0].scrollIntoView();
    });
  }

  return {
    load: load,
  }
}

function NewLoadingIndicator(opts) {
  var loading = $('<div class="loading"></div>').appendTo($('body'));
  var lastReqId = -1;

  function on(reqId) {
    if (lastReqId < reqId) {
      lastReqId = reqId;
      loading.text(opts.on);
    }
  }

  function off(reqId) {
    if (lastReqId <= reqId) {
      lastReqId = reqId;
      loading.text(opts.off);
    }
  }

  return {
    on: on,
    off: off
  }
}

function NewBlocker(el) {
  return {
    on: function() { el.show(); },
    off: function() { el.hide(); }
  }
}

function NewExampleLocation() {
  var prefix = "#example:";
  return {
    isSet: function() {
      return window.location.hash && window.location.hash.startsWith(prefix);
    },
    get: function() {
      var defaultExample = "example-demo";
      if (window.location.hash) {
        if (window.location.hash.startsWith(prefix)) {
          defaultExample = window.location.hash.replace(prefix, "", 1)
        }
      }
      return defaultExample
    },
    set: function(example) {
      window.location.hash = "#example:"+example
    }
  }
}

function NewGistLocation() {
  var prefix = "#gist:";
  return {
    isSet: function() {
      return window.location.hash && window.location.hash.startsWith(prefix);
    },
    get: function() {
      var id = null;
      if (window.location.hash) {
        if (window.location.hash.startsWith(prefix)) {
          id = window.location.hash.replace(prefix, "", 1);
        }
      }
      // Gist URL looks like this: https://gist.github.com/cppforlife/ed41c0e8ab19029347732ba247a6c460
      if (id && id.startsWith("https://")) {
        id = id.split("/").slice(-1)[0];
      }
      return id
    },
    set: function(id) {
      window.location.hash = prefix+id;
    }
  }
}

function NewGistForm(formEl, gistLocation, gist) {
  formEl.submit(function(event) {
    event.preventDefault();
    gistLocation.set($("input", formEl).val());
    console.log(gistLocation.get());
    gist.load(gistLocation.get(), {
      scrollIntoView: true,
      preDoneCallback: function(exampleId) {
        $("#playground").show();
      }
    });
    return false;
  });

  return {}
}

function NewGoogleAnalytics() {
  return {
    recordExampleClick: function(name) {
      if (!window.ga) return;
      window.ga('send', {
        hitType: 'event',
        eventCategory: 'example',
        eventAction: 'show',
        eventLabel: name,
      });
    }
  }
}

$(document).ready(function() {
  var templatesLoadingIndicator = NewLoadingIndicator({
    on: "Playground: templating... in progress",
    off: "Playground: templating... done"
  });
  var templates = NewTemplates($("#playground"), {
    preEvaluateCallback: templatesLoadingIndicator.on,
    postEvaluateCallback: templatesLoadingIndicator.off,
  });

  var googAnalytics = NewGoogleAnalytics();
  var uiBlocker = NewBlocker($("#playground .blocker"));

  var exampleLocation = NewExampleLocation();
  var examples = NewExamples($("#playground"),
    templates, exampleLocation, uiBlocker);

  var gistLocation = NewGistLocation();
  var gist = NewGist($("#playground"),
    templates, gistLocation, uiBlocker);
  var gistForm = NewGistForm($("#show-gist"), gistLocation, gist);

  if (gistLocation.isSet()) {
    gist.load(gistLocation.get(), {
      scrollIntoView: true,
      preDoneCallback: function(exampleId) {
        $("#playground").show();
      }
    });
  } else {
    examples.load(exampleLocation.get(), {
      scrollIntoView: exampleLocation.isSet(),
      preDoneCallback: function(exampleId) {
        $("#playground").show();
        googAnalytics.recordExampleClick(exampleId);
      }
    });
  }

  $("#playground .add-config").click(function() {
    templates.addFile(null, {focus: true});
    templates.evaluate();
    return false;
  });

  $("#playground .run").click(function(){
    templates.evaluate();
    return false;
  })

  $("#playground .expand").click(function(){
    $("body").toggleClass("playground-expanded");
    $("#playground")[0].scrollIntoView();
    return false;
  })
})
