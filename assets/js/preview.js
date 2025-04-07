document.addEventListener('DOMContentLoaded', () => {
	const previewName = document.getElementById('preview-name');
	const previewIdNumber = document.getElementById('preview-id-number');
	const previewBirthDate = document.getElementById('preview-birth-date');
	const previewAddress = document.getElementById('preview-address');

	const path = window.location.pathname;
	if (path.includes('/preview/stages/2/')) {
		previewName.textContent = `邱吉爾`;
		previewIdNumber.textContent = `A123456789`;
		if (typeof dateDelimiter === 'string' && dateDelimiter !== "") {
			previewBirthDate.textContent = `888${dateDelimiter}11${dateDelimiter}30`;
		} else {
			previewBirthDate.textContent = `888 年 11 月 30 日`;
		}
		previewAddress.textContent = `某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址`;
	} else {
		const data = JSON.parse(sessionStorage.getItem('previewData'));
		if (!data) return;
		sessionStorage.removeItem('previewData');

		previewName.textContent = data.name;
		previewIdNumber.textContent = data.idNumber;
		if (dateDelimiter !== "") {
			previewBirthDate.textContent = `${data.birthYear}${dateDelimiter}${data.birthMonth}${dateDelimiter}${data.birthDay}`;
		} else {
			previewBirthDate.textContent = `${data.birthYear} 年 ${data.birthMonth} 月 ${data.birthDay} 日`;
		}
		previewAddress.textContent = data.address;
	}
});
