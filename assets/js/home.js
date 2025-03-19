const filteredCandidateContainer = document.getElementById('filtered-candidate-container');
const shareContainer = document.getElementById('share-container');
const pepTalk = document.querySelector(`.pep-talk`);

async function sendAjaxRequest(municipality, district, ward) {
	let params = new URLSearchParams();

	if (municipality !== null && municipality !== undefined) {
		params.append("municipality", municipality);
	}
	if (district !== null && district !== undefined) {
		params.append("district", district);
	}
	if (ward !== null && ward !== undefined) {
		params.append("ward", ward);
	}

	let fullUrl = `${baseURL}/apis/constituencies?${params.toString()}`;

	try {
		let response = await fetch(fullUrl, { method: "GET" });

		if (!response.ok && response.status !== 404) {
			let error = new Error(`HTTP Error: ${response.status}`);
			error.status = response.status;
			throw error;
		}

		return await response.json();
	} catch (error) {
		throw error;
	}
}

document.addEventListener("DOMContentLoaded", () => {
	const municipalitiesSelect = document.getElementById("filter-municipalities");
	const districtsSelect = document.getElementById("filter-districts");
	const wardsSelect = document.getElementById("filter-wards");
	const initMunicipalityHasFailed = (document.querySelector(`.municipalities ul[data-city="1"] li.recall-failed`) !== null) ? true : false;
	if (initMunicipalityHasFailed) {
		pepTalk.style.display = "flex";
	}

	const resetSelect = (select) => {
		const defaultOption = select.querySelector('option[value=""]');
		select.disabled = true;
		select.innerHTML = "";
		if (defaultOption) {
			select.appendChild(defaultOption.cloneNode(true));
		}
	};

	const populateOptions = (select, divisions) => {
		divisions.forEach(division => {
			const elem = document.createElement("option");
			elem.value = division.id;
			elem.textContent = division.n;
			select.appendChild(elem);
		});
		select.disabled = false;
	};

	municipalitiesSelect.addEventListener("change", () => {
		mask.classList.add('active');
		filteredCandidateContainer.innerHTML = "";
		shareContainer.style.display = "none";
		resetSelect(districtsSelect);
		resetSelect(wardsSelect);

		sendAjaxRequest(municipalitiesSelect.value, null, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					populateOptions(districtsSelect, data.result.divisions);
				} else {
					console.error("invalid municipality");
				}
			})
			.catch(error => {
				console.error(error);
			})
			.finally(() => {
				mask.classList.remove('active');
			});
	});

	districtsSelect.addEventListener("change", () => {
		mask.classList.add('active');
		filteredCandidateContainer.innerHTML = "";
		shareContainer.style.display = "none";
		resetSelect(wardsSelect);

		sendAjaxRequest(municipalitiesSelect.value, districtsSelect.value, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					populateOptions(wardsSelect, data.result.divisions);
				} else {
					console.error("invalid district");
				}
			})
			.catch(error => {
				console.error(error);
			})
			.finally(() => {
				mask.classList.remove('active');
			});
	});

	wardsSelect.addEventListener("change", () => {
		mask.classList.add('active');
		filteredCandidateContainer.innerHTML = "";
		shareContainer.style.display = "none";
		sendAjaxRequest(municipalitiesSelect.value, districtsSelect.value, wardsSelect.value)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "legislators")) {
					const address = municipalitiesSelect.selectedOptions[0].text + districtsSelect.selectedOptions[0].text + wardsSelect.selectedOptions[0].text;
					showFilteredCandidateContainer(data.result.legislators, address);
				} else {
					console.error("invalid ward");
				}
			})
			.catch(error => {
				console.error(error);
			})
			.finally(() => {
				mask.classList.remove('active');
			});
	});
	
	dialogMask.addEventListener("click", function(event) {
		if (event.target === dialogClose || dialogClose.contains(event.target)) {
			dialogMask.style.display = "none";
			return;
		}

		if (!dialog.contains(event.target)) {
			dialogMask.style.display = "none";
		}
	});
});

function showFilteredCandidateContainer(legislators, address) {
	if (Array.isArray(legislators)) {
		legislators.forEach(legislator => {
			const candidateContainer = document.createElement("div");
			candidateContainer.classList.add("candidate-container");

			let candidateAction = "",
				recallStages = "",
				recallFailedClass = "";

			switch (legislator.recallStatus) {
				case "ABORTED":
					recallFailedClass = "recall-failed";
					recallStages = `<div class="recall-stage-failed-flow">
						<h4>您選區的連署未能及時送件...</h4>
						別灰心，我們還是需要您的力量，支持其他選區進行中的罷免活動，幫忙分享資訊！
					</div>`;
					candidateAction = `<button class="btn-black lg w100" onclick="shareCurrentLink('')"><i class="icon-link"></i>幫忙分享資訊！</button>`;
					break;

				case "FAILED":
					recallFailedClass = "recall-failed";
					recallStages = `<div class="recall-stage-failed-flow">
						<h4>您選區的連署未通過...</h4>
						別灰心，我們還是需要您的力量，支持其他選區進行中的罷免活動，幫忙分享資訊！
					</div>`;
					candidateAction = `<button class="btn-black lg w100" onclick="shareCurrentLink('')"><i class="icon-link"></i>幫忙分享資訊！</button>`;
					break;

				default:
					recallStages = [1, 2, 3].map(stage => `
						<h4 class="recall-stage ${stage === legislator.recallStage ? 'active' : ''}">
							<span>第 ${stage} 階段</span>${stage === 3 ? "罷免投票" : "連署罷免"}
						</h4>
						${stage < 3 ? '<span class="icon-step-arrow"></span>' : ''}
					`).join('');

					recallStages = `<div class="recall-stage-flow">${recallStages}</div>`;

					if (legislator.recallStage === 1 || legislator.recallStage === 2) {
						if (legislator.formDeployed) {
							candidateAction = `<a href="${legislator.participateURL}?address=${address}"><button class="btn-primary lg">連署罷免</button></a>`;
						} else {
							candidateAction = `<button class="btn-primary lg" disabled>${legislator.recallStage} 階準備中</button>`
						}
					} else {
						candidateAction = `<a href="${legislator.calendarURL}" target="_blank"><button class="btn-primary lg w100">加入 Google 日曆提醒投票</button></a>`;
					}
					break;
			}

			candidateContainer.innerHTML = `
				<div class="candidate ${recallFailedClass}">
					<div class="candidate-name">${legislator.politicianName}</div>
					<div class="candidate-zone">${legislator.constituencyName}</div>
				</div>
				<div class="recall-stage-container">
					${recallStages}
					<div class="candidate-action">
						${candidateAction}
					</div>
				</div>
				${legislator.recallStatus === "ONGOING" ? "<p>罷免需經兩個階段連署，兩階段都通過後才進行投票決定罷免結果。請大家務必三個階段都完整參與！</p>" : ''}`;
			filteredCandidateContainer.appendChild(candidateContainer);
		});
		filteredCandidateContainer.style.display = "flex";
		filteredCandidateContainer.scrollIntoView({ behavior: "smooth" });
	}
}

function showShareContainer() {
	shareContainer.style.display = "block";
	shareContainer.scrollIntoView({ behavior: "smooth" });
}

new Swiper('.swiper', {
	slidesPerView: 1.05,
	spaceBetween: 16,
	pagination: { el: ".swiper-pagination", clickable: true },
	autoplay: {
		delay: 3000,
		disableOnInteraction: false
	},
	loop: true
});

const municipalityLists = document.querySelectorAll(".municipalities ul");
const municipalityTags = document.querySelectorAll(".municipality-tag");

function toggleCityList(cityId) {
	const targetUl = document.querySelector(`ul[data-city="${cityId}"]`);
	const targetTag = document.querySelector(`.municipality-tag[data-city="${cityId}"]`);
	const hasFailed = (targetUl.querySelector(`li.recall-failed`) !== null) ? true : false;

	if (targetUl && targetUl.style.display === "flex") {
		return;
	}

	municipalityLists.forEach(ul => (ul.style.display = "none"));

	municipalityTags.forEach(tag => {
		tag.classList.remove("active");
		tag.firstElementChild.classList.remove("active");
	});

	if (targetUl) {
		targetUl.style.display = "flex";
	}

	if (targetTag) {
		targetTag.classList.add("active");
		targetTag.firstElementChild.classList.add("active");
	}

	if (hasFailed) {
		pepTalk.style.display = "flex";
	} else {
		pepTalk.style.display = "none";
	}

	targetUl.scrollIntoView({ behavior: "smooth" });
}
