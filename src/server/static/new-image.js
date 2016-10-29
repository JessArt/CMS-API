$(function() {
  $('#tags').select2({
    tags: true
  });

  $('.fancybox').fancybox();

  var uploadInput = document.getElementById('upload');
  $(uploadInput).on('change', function() {
    if (uploadInput.files && uploadInput.files[0]) {
      var reader = new FileReader();

      reader.onload = function (e) {
        $('#uploaded-image').attr('src', e.target.result);
      }

      reader.readAsDataURL(uploadInput.files[0]);
    }
  });
});
