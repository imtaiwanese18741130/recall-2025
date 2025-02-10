async function downloadPDF(filename, redirectURL) {
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
				scale: 2,
				ignoreElements: (elem) => elem.classList.contains('whereToSign'),
			});

			const imgData = canvas.toDataURL('image/webp', 1.0);
			const pdfHeight = (canvas.height * pdfWidth) / canvas.width;

			pdf.addImage(imgData, 'PNG', 0, 0, pdfWidth, pdfHeight);
		}

		pdf.save(`${filename}.pdf`);
		window.location.href = redirectURL;
	} catch (error) {
		console.error('Error generating PDF:', error);
	} finally {
		mask.classList.remove('active');
	}
}
async function copyLink(url) {
	navigator.clipboard.writeText(url)
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
			console.error('cannot copy:', err);
		});
}
async function copyCurrentLink() {
	const currentUrl = window.location.href;
	copyLink(currentUrl)
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
