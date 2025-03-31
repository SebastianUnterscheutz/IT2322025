function handleOfferFormSubmit(event) {
    event.preventDefault();
    alert('Mitfahrgelegenheit muss per E-Mail aktiviert werden.');
    window.location.href = 'index.html';
}

document.addEventListener('DOMContentLoaded', function() {
    const offerForm = document.getElementById('offerForm');
    if (offerForm) {
        offerForm.addEventListener('submit', handleOfferFormSubmit);
    }
});