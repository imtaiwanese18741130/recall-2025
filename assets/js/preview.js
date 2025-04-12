const arabicToChinese = {
  "1": "一",
  "2": "二",
  "3": "三",
  "4": "四",
  "5": "五",
  "6": "六",
  "7": "七",
  "8": "八",
  "9": "九",
  "10": "十"
};

function convertLargeNumber(num) {
  if (num.length === 1) {
    return arabicToChinese[num] || "";
  } else if (num.length === 2) {
    const digits = num.split("");
    let result = "";
    if (digits[0] === "1") {
      result += "十";
    } else {
      result += (arabicToChinese[digits[0]] || "") + "十";
    }
    if (digits[1] !== "0") {
      result += arabicToChinese[digits[1]] || "";
    }
    return result;
  }
  return "";
}

function sanitizeAddress(address) {
	address = address.replace(/\s+/g, "");
  address = address.replace(/-/g, "之");

  const pattern = /(\d+)(段|樓)/g;

  return address.replace(pattern, (match, num, unit) => {
    if (arabicToChinese[num]) {
      return arabicToChinese[num] + unit;
    }
    const converted = convertLargeNumber(num);
    if (converted) {
      return converted + unit;
    }
    return match;
  });
}

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
		previewAddress.textContent = sanitizeAddress(data.address);
	}
});

