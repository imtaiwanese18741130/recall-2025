const filteredCandidateContainer = document.getElementById('filtered-candidate-container');
const shareContainer = document.getElementById('share-container');
const pepTalk = document.querySelector(`.pep-talk`);

const NewSearchCandidateHandler = (config) => {
	const layers = {
		municipality: {
			title: "縣市",
			input: config.filterMunicipalitiesInput,
			ul: config.filterMunicipalitiesUl,
			wrapper: config.filterMunicipalitiesWrapper,
			backBtn: config.filterMunicipalitiesBackBtn,
			data: [],
			selected: null,
		},
		district: {
			title: "行政區",
			input: config.filterDistrictsInput,
			ul: config.filterDistrictsUl,
			wrapper: config.filterDistrictsWrapper,
			backBtn: config.filterDistrictsBackBtn,
			data: [],
			selected: null,
		},
		ward: {
			title: "鄉鎮村里",
			input: config.filterWardsInput,
			ul: config.filterWardsUl,
			wrapper: config.filterWardsWrapper,
			backBtn: config.filterWardsBackBtn,
			data: [],
			selected: null,
		},
	};

	function bindFilterEvents (level, selectFunc, goBackFunc) {
		const { input, ul, wrapper, backBtn } = layers[level];

		input.addEventListener('focus', (e) => {
			// e.preventDefault();
			// setTimeout(() => {
			// 	input.focus();
			// }, 1000);

			populateFilter(input, ul, filterOptions(level, ""), selectFunc, wrapper);


			// if (window.innerWidth <= 600) {
			// 	setTimeout(() => {
			// 		console.log("scrollIntoView");
			// 		input.scrollIntoView(true);
			// 	}, 1000);
			// }

			// if (window.visualViewport) {
			// 	console.log("visualViewport");
			// 	const viewportHeight = window.visualViewport.height;
			// 	const offsetTop = window.visualViewport.offsetTop;
			// 	const filterMunicipalitiesUl = document.getElementById('filter-municipalities-ul');
			// 	filterMunicipalitiesUl.style.height = `${viewportHeight-125}px`;
			// 	filterMunicipalitiesUl.style.top = `${offsetTop}px`;
			// }
		});
		input.addEventListener('input', (e) => {
			populateFilter(input, ul, filterOptions(level, e.target.value.trim()), selectFunc, wrapper);
		});
		backBtn.addEventListener('click', () => {
			goBackFunc(true);
			wrapper.classList.remove("open");
		});

		document.addEventListener('click', (e) => {
			if (!wrapper.contains(e.target)) goBackFunc(false);
		});
	};

	function generateSynonymVariants(str) {
		const variants = new Set([str]);
		const synonymMap = {
			"台": ["臺"],
			"臺": ["台"],
		};

		for (const [key, values] of Object.entries(synonymMap)) {
			if (str.includes(key)) {
				values.forEach(val => {
					variants.add(str.replace(new RegExp(key, "g"), val));
				});
			}
		}

		return Array.from(variants);
	}

	function filterOptions (level, searchTerm) {
		const opts = layers[level].data;

		if (opts.length === 0) return [];
		if (searchTerm === "") return opts;

		const searchTermVariants = generateSynonymVariants(searchTerm.toLowerCase());

		filteredOpts = opts.filter(opt => {
			const optNameVariants = generateSynonymVariants(opt.n.toLowerCase());

			return searchTermVariants.some(searchVariant =>
				optNameVariants.some(optVariant =>
					optVariant.includes(searchVariant)
				)
			);
		});

		if (filteredOpts.length === 0) {
			return [
				{
					id: null,
					n: `請輸入正確${layers[level].title}`
				}
			];
		}

		return filteredOpts
	};

	function populateFilter(inputElem, ulElem, opts, onClick, wrapperElem) {
		ulElem.innerHTML = '';

		opts.forEach(opt => {
			const li = document.createElement('li');
			li.value = opt.id;
			li.textContent = opt.n;

			if (opt.id !== null) {
				li.addEventListener('click', () => {
					inputElem.value = opt.n;
					ulElem.style.display = 'none';
					onClick(opt);

					if (wrapperElem) {
						wrapperElem.classList.remove("open");
					}
				});

				li.addEventListener('mouseover', () => {
					li.style.backgroundColor = '#f0f0f0';
				});
				li.addEventListener('mouseout', () => {
					li.style.backgroundColor = '';
				});
			}

			ulElem.appendChild(li);
		});

		ulElem.style.display = opts.length ? 'block' : 'none';

		if (wrapperElem && opts.length) {
			setTimeout(() => wrapperElem.classList.add("open"), 10);
		} else if (wrapperElem) {
			wrapperElem.classList.remove("open");
		}
	};

	return {
		init() {
			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendListMunicipalitiesReq()
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					this.setMunicipalities(data.result.divisions);
					this.resetMunicipalityFilter();
				} else {
					console.error("invalid district");
				}
			})
			.catch(error => console.error(error))
			.finally(() => mask.classList.remove('active'));

			bindFilterEvents('municipality', this.selectMunicipality.bind(this), this.goBackMunicipality.bind(this));
			bindFilterEvents('district', this.selectDistrict.bind(this), this.goBackDistrict.bind(this));
			bindFilterEvents('ward', this.selectWard.bind(this), this.goBackWard.bind(this));
		},

		selectMunicipality(municipality) {
			layers.municipality.selected = municipality;

			this.resetDistrictFilter(false);
			this.resetWardFilter(true);

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(layers.municipality.selected.id, null, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					this.setDistricts(data.result.divisions);
				} else {
					console.error("invalid municipality");
				}
			})
			.catch(error => console.error(error))
			.finally(() => mask.classList.remove('active'));
		},
		selectDistrict(district) {
			layers.district.selected = district;

			this.resetWardFilter(false);

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(layers.municipality.selected.id, layers.district.selected.id, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					this.setWards(data.result.divisions);
				} else {
					console.error("invalid district");
				}
			})
			.catch(error => console.error(error))
			.finally(() => mask.classList.remove('active'));
		},
		selectWard(ward) {
			layers.ward.selected = ward;

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(layers.municipality.selected.id, layers.district.selected.id, layers.ward.selected.id)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "legislators")) {
					const address = `${layers.municipality.selected.n}${layers.district.selected.n}${layers.ward.selected.n}`;
					showFilteredCandidateContainer(data.result.legislators, address);
				} else {
					console.error("invalid ward");
				}
			})
			.catch(error => console.error(error))
			.finally(() => mask.classList.remove('active'));
		},

		// Municipalities, Districts, Wards Reset
		resetMunicipalityFilter() {
			layers.municipality.input.disabled = false;
			layers.municipality.input.value = "";
			layers.municipality.ul.innerHTML = '';
			layers.municipality.ul.style.display = 'none';

			this.resetDistrictFilter(true);
			this.resetWardFilter(true);
		},
		resetDistrictFilter(inputDisabled) {
			layers.district.selected = null;

			layers.district.input.disabled = inputDisabled;
			layers.district.input.value = "";
			layers.district.ul.innerHTML = '';
			layers.district.ul.style.display = 'none';

			this.resetWardFilter(true);
		},
		resetWardFilter(inputDisabled) {
			layers.ward.selected = null;

			layers.ward.input.disabled = inputDisabled;
			layers.ward.input.value = "";
			layers.ward.ul.innerHTML = '';
			layers.ward.ul.style.display = 'none';
		},

		// Municipalities, Districts, Wards GoBack
		goBackMunicipality(forGoBack) {
			if (forGoBack) layers.municipality.input.disabled = false;
			layers.municipality.input.value = (layers.municipality.selected) ? layers.municipality.selected.n : "";
			layers.municipality.ul.innerHTML = '';
			layers.municipality.ul.style.display = 'none';
		},
		goBackDistrict(forGoBack) {
			if (forGoBack) layers.district.input.disabled = (layers.district.data.length > 0) ? false : true;
			layers.district.input.value = (layers.district.selected) ? layers.district.selected.n : "";
			layers.district.ul.innerHTML = '';
			layers.district.ul.style.display = 'none';
		},
		goBackWard(forGoBack) {
			if (forGoBack) layers.ward.input.disabled = (layers.ward.data.length > 0) ? false : true;
			layers.ward.input.value = (layers.ward.selected) ? layers.ward.selected.n : "";
			layers.ward.ul.innerHTML = '';
			layers.ward.ul.style.display = 'none';
		},

		// Municipalities, Districts, Wards Setters
		setMunicipalities(municipalities) {
			layers.municipality.data = municipalities;
		},
		setDistricts (districts) {
			layers.district.data = districts;
		},
		setWards(wards) {
			layers.ward.data = wards;
		},
	};
}

async function sendListMunicipalitiesReq() {
	let fullUrl = `${baseURL}/apis/municipalities`;

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

async function sendSearchConstituenciesReq(municipality, district, ward) {
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
	const initMunicipalityHasFailed = (document.querySelector(`.municipalities ul[data-city="1"] li.recall-failed`) !== null) ? true : false;
	if (initMunicipalityHasFailed) {
		pepTalk.style.display = "flex";
	}

	const searchHdr = NewSearchCandidateHandler(
		{
			filterMunicipalitiesWrapper: document.getElementById("municipalities-filter-wrapper"),
			filterMunicipalitiesInput: document.getElementById("filter-municipalities-input"),
			filterMunicipalitiesUl: document.getElementById("filter-municipalities-ul"),
			filterMunicipalitiesBackBtn: document.getElementById("filter-municipalities-back-btn"),

			filterDistrictsWrapper: document.getElementById("districts-filter-wrapper"),
			filterDistrictsInput: document.getElementById("filter-districts-input"),
			filterDistrictsUl: document.getElementById("filter-districts-ul"),
			filterDistrictsBackBtn: document.getElementById("filter-districts-back-btn"),

			filterWardsWrapper: document.getElementById("wards-filter-wrapper"),
			filterWardsInput: document.getElementById("filter-wards-input"),
			filterWardsUl: document.getElementById("filter-wards-ul"),
			filterWardsBackBtn: document.getElementById("filter-wards-back-btn"),
		}
	);
	searchHdr.init();

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
						<div class="urgency">
							<div class="days-left">
								<i class="icon-urgent"></i>
								${legislator.daysLeft < 15 ? '<i class="icon-urgent"></i>' : ''} 
								${legislator.daysLeft < 0 ? '<i class="icon-urgent"></i>' : ''} 
								${legislator.daysLeft > 0
									? `${legislator.safetyCutoffDateStr} 截止，剩餘 ${legislator.daysLeft} 天`
									: `請儘速繳交，罷團已開始造冊`}
							</div>
						</div>
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