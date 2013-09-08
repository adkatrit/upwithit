$('document').ready(function(){

	$('input').keypress(function (e) {
		
	  if (e.which == 13) {
	  	e.preventDefault()
	    var query = $(this).val().trim()
	    $.get('http://localhost:8333/',function(response){

	    	console.log(response)
	    });
	    
	  }
	});
});