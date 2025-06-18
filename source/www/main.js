document.addEventListener("DOMContentLoaded", function () {
    var form = document.getElementById("uploadForm");
    var submit = document.getElementById("submit");
    var asciiOutput = document.getElementById("asciiOutput");

    if (form && submit && asciiOutput) {
        form.addEventListener("submit", function (event) {
            event.preventDefault();
            submit.disabled = true;
            asciiOutput.textContent = "Uploading...";
            var formData = new FormData(form);
            fetch("/upload", {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) throw new Error(response.statusText);
                return response.text();
            })
            .then(text => {
                asciiOutput.textContent = text;
            })
            .catch(error => {
                asciiOutput.textContent = "Error: " + error.message;
            })
            .finally(() => {
                submit.disabled = false;
            });
        });
    }

    // Handle the banner submission 
    var bannerForm = document.getElementById("bannerGen");
    var bannerSubmit = document.getElementById("bannerSubmit");
    var asciiOutput = document.getElementById("asciiOutput");
    if (bannerForm && bannerSubmit && asciiOutput) {
        bannerForm.addEventListener("submit", function (event) {
            event.preventDefault();
            bannerSubmit.disabled = true;
            asciiOutput.textContent = "Generating banner...";
            var formData = new FormData(bannerForm);
            fetch("/banner", {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) throw new Error(response.statusText);
                return response.text();
            })
            .then(text => {
                asciiOutput.textContent = text;
            })
            .catch(error => {
                asciiOutput.textContent = "Error: " + error.message;
            })
            .finally(() => {
                bannerSubmit.disabled = false;
            });
        });
    }
});