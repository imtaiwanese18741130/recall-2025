async function preparePDF(filename, redirectURL, isDownload) {
	const { jsPDF } = window.jspdf;
	const containers = document.querySelectorAll('.a4-portrait, .a4-landscape');
	if (containers.length === 0) {
		console.error('No elements found with the given class name.');
		return;
	}
	const mask = document.querySelector('.mask');
	mask.classList.add('active');

	try {
		let pdf;
		for (const [index, container] of containers.entries()) {
			const isLandscape = container.classList.contains('a4-landscape');
			const orientation = isLandscape ? 'l' : 'p';

			if (index === 0) {
				pdf = new jsPDF({ orientation, unit: 'mm', format: 'a4' });
			} else {
				pdf.addPage('a4', orientation);
			}

			const pdfWidth = pdf.internal.pageSize.getWidth();
			const canvas = await html2canvas(container, {
				scale: 1.5,
				ignoreElements: (elem) => elem.classList.contains('whereToSign'),
			});
			const imgData = canvas.toDataURL('image/png', 1.0);
			const pdfHeight = (canvas.height * pdfWidth) / canvas.width;
			pdf.addImage(imgData, 'PNG', 0, 0, pdfWidth, pdfHeight);
		}

		if (isDownload) {
			await downloadPDF(pdf, filename);
		} else {
			await previewPDF(pdf);
		}

		window.location.href = redirectURL;
	} catch (error) {
		console.error('Error generating PDF:', error);
	} finally {
		mask.classList.remove('active');
	}
}

async function downloadPDF(pdf, filename) {
	await new Promise((resolve) => {
		pdf.save(`${filename}.pdf`, { returnPromise: false });
		setTimeout(resolve, 500);
	});
}

async function previewPDF(pdf) {
	await new Promise((resolve) => {
		const url = window.URL.createObjectURL(pdf.output('blob'));
		const newWindow = window.open(url, '_blank');

		if (newWindow) {
			const checkWindowClosed = setInterval(() => {
				if (newWindow.closed) {
					window.URL.revokeObjectURL(url);
					clearInterval(checkWindowClosed);
				}
			}, 1000);
		}
		setTimeout(resolve, 500);
	});
}

async function shareLink(text, url) {
	try {
		await navigator.share({title: "守護我們珍愛的臺灣，我們需要你！", text: text, url: url});
	} catch (err) {
		console.error("share failed:", err);
	}
}

async function shareCurrentLink(text) {
	shareLink(text, window.location.href)
}

async function copyInnerText(containerSelector) {
	const container = document.querySelector(containerSelector);

	if (!container) {
		console.error(`cannot find container: ${containerSelector}`);
		return;
	}

	navigator.clipboard.writeText(container.innerText)
		.then(() => {
			const popout = document.getElementById('popout');
			popout.classList.remove('hidden');
			popout.classList.add('visible');

			setTimeout(() => {
				popout.classList.remove('visible');
				popout.classList.add('hidden');
			}, 2000);
		})
		.catch(err => {
			console.error('cannot copy', err);
		});
}

function isValidIdNumber(idNumber) {
	if (!/^[A-Z][12]\d{8}$/.test(idNumber)) {
		return false;
	}

	const letterMap = {
		A: 10, B: 11, C: 12, D: 13, E: 14,
		F: 15, G: 16, H: 17, J: 18, K: 19,
		L: 20, M: 21, N: 22, P: 23, Q: 24,
		R: 25, S: 26, T: 27, U: 28, V: 29,
		X: 30, Y: 31, W: 32, Z: 33, I: 34,
		O: 35,
	};

	const firstLetter = idNumber[0];
	const firstDigit = letterMap[firstLetter];

	if (!firstDigit) {
		return false;
	}

	const digits = [
		Math.floor(firstDigit / 10),
		firstDigit % 10,
	];

	for (let i = 1; i < 10; i++) {
		const num = parseInt(idNumber[i], 10);
		if (isNaN(num)) {
			return false;
		}
		digits.push(num);
	}

	const checksum =
		digits[0] +
		digits[1] * 9 +
		digits[2] * 8 +
		digits[3] * 7 +
		digits[4] * 6 +
		digits[5] * 5 +
		digits[6] * 4 +
		digits[7] * 3 +
		digits[8] * 2 +
		digits[9] +
		digits[10];

	return checksum % 10 === 0;
}

function generateQRCode(data, size) {
	const qrcodeContainer = document.getElementById("qrcode-container");
	const qrcode = qrcodeContainer.querySelector(".qrcode");
	qrcode.innerHTML = "";
	new QRCode(qrcode, {
		text: data,
		width: size,
		height: size,
		correctLevel: QRCode.CorrectLevel.H
	});
	const logo = qrcodeContainer.querySelector(".qr-logo");
	logo.innerHTML = "OurTaiwan<br>製作連署書";
	logo.style.fontSize = `${size * 0.075}px`;
	logo.style.padding = `${size * 0.01}px ${size * 0.025}px`;
}

function isValidDate(year, month, day) {
	if (month < 1 || month > 12 || day < 1 || day > 31) {
		return false;
	}
	const fullYear = 1911 + parseInt(year, 10);
	const date = new Date(fullYear, month - 1, day);
	return date.getFullYear() === fullYear && date.getMonth() === month - 1 && date.getDate() === day;
}

async function downloadQRCode(data) {
  try {
    const size = 4096;
    const dlQRCodeContainer = document.createElement("div");
    dlQRCodeContainer.classList.add("downloaded-qrcode-container");
    dlQRCodeContainer.innerHTML = `<div class="qrcode"></div><div class="qr-logo"></div>`;
    document.body.appendChild(dlQRCodeContainer);

    const dlQRCode = dlQRCodeContainer.querySelector(".qrcode");
    new QRCode(dlQRCode, {
      text: data,
      width: size,
      height: size,
      correctLevel: QRCode.CorrectLevel.H
    });

    const dlLogo = dlQRCodeContainer.querySelector(".qr-logo");
    dlLogo.innerHTML = "OurTaiwan<br>製作連署書";
    dlLogo.style.fontSize = `300px`;
    dlLogo.style.padding = `24px 24px`;
    dlLogo.style.borderRadius = `64px`;

    const canvas = await html2canvas(dlQRCodeContainer, { backgroundColor: "#ffffff", scale: 1 });
    const blob = await new Promise((resolve) => canvas.toBlob(resolve, "image/png"));

    if (!blob) throw new Error("Failed to create Blob from canvas");

    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    link.download = "qrcode.png";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(link.href);

    document.body.removeChild(dlQRCodeContainer);
  } catch (error) {
    console.error("Error generating QR code:", error);
  }
}
