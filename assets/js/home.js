const filteredCandidateContainer = document.getElementById('filtered-candidate-container');
const shareContainer = document.getElementById('share-container');
const pepTalk = document.querySelector(`.pep-talk`);

const NewSearchCandidateHandler = (config) => {
	const filterMunicipalitiesWrapper = config.filterMunicipalitiesWrapper;
	const filterMunicipalitiesInput = config.filterMunicipalitiesInput;
	const filterMunicipalitiesUl = config.filterMunicipalitiesUl;

	const filterDistrictsWrapper = config.filterDistrictsWrapper;
	const filterDistrictsInput = config.filterDistrictsInput;
	const filterDistrictsUl = config.filterDistrictsUl;

	const filterWardsWrapper = config.filterWardsWrapper;
	const filterWardsInput = config.filterWardsInput;
	const filterWardsUl = config.filterWardsUl;

	let _currDistricts = [];
	let _currWards = [];

	let _selectMunicipality = null;
	let _selectDistrict = null;
	let _selectWard = null;

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

			// municipalities
			document.addEventListener('click', (e) => {
				if (!filterMunicipalitiesWrapper.contains(e.target)) {
					filterMunicipalitiesUl.style.display = 'none';
				}
			});
			filterMunicipalitiesInput.addEventListener('focus', () => {
				this.populateFilter(filterMunicipalitiesInput, filterMunicipalitiesUl,
					_currDistricts,
					(municipality) => {
						this.selectMunicipality(municipality);
					},
				);
			});
			filterMunicipalitiesInput.addEventListener('input', (e) => {
				this.populateFilter(filterMunicipalitiesInput, filterMunicipalitiesUl,
					_currDistricts.filter(municipality => municipality.n.includes(e.target.value.trim())),
					(municipality) => {
						this.selectMunicipality(municipality);
					},
				);
			});

			// districts
			document.addEventListener('click', (e) => {
				if (!filterDistrictsWrapper.contains(e.target)) {
					filterDistrictsUl.style.display = 'none';
				}
			});
			filterDistrictsInput.addEventListener('focus', () => {
				this.populateFilter(filterDistrictsInput, filterDistrictsUl,
					this.fitlterDistricts(""),
					(district) => {
						this.selectDistrict(district);
					},
				);
			});
			filterDistrictsInput.addEventListener('input', (e) => {
				this.populateFilter(filterDistrictsInput, filterDistrictsUl,
					this.fitlterDistricts(e.target.value.trim()),
					(district) => {
						this.selectDistrict(district);
					},
				);
			});

			// wards
			document.addEventListener('click', (e) => {
				if (!filterWardsWrapper.contains(e.target)) {
					filterWardsUl.style.display = 'none';
				}
			});
			filterWardsInput.addEventListener('focus', () => {
				this.populateFilter(filterWardsInput, filterWardsUl,
					this.fitlterWards(""),
					(ward) => {
						this.selectWard(ward);
					},
				);
			});
			filterWardsInput.addEventListener('input', (e) => {
				this.populateFilter(filterWardsInput, filterWardsUl,
					this.fitlterWards(e.target.value.trim()),
					(ward) => {
						this.selectWard(ward);
					},
				);
			});
		},

		selectMunicipality(municipality) {
			_selectMunicipality = municipality;
			this.resetDistrictFilter(false);
			this.resetWardFilter(true);

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(municipality.id, null, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					this.setDistricts(data.result.divisions);
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
		},
		selectDistrict(district) {
			_selectDistrict = district;

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(_selectMunicipality.id, district.id, null)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "divisions")) {
					this.setWards(data.result.divisions);
					this.resetWardFilter(false);
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
		},
		selectWard(ward) {
			_selectWard = ward;

			mask.classList.add('active');
			shareContainer.style.display = "none";
			sendSearchConstituenciesReq(_selectMunicipality.id, _selectDistrict.id, ward.id)
			.then(data => {
				if (!Object.hasOwn(data, "result")) {
					showShareContainer();
				} else if (Object.hasOwn(data.result, "legislators")) {
					const address = _selectMunicipality.n + _selectDistrict.n + ward.n;
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
		},

		resetDistrictFilter(inputDisabled) {
			_selectDistrict = null;

			filterDistrictsInput.disabled = inputDisabled;
			filterDistrictsInput.value = "";
			filterDistrictsUl.innerHTML = '';
			filterDistrictsUl.style.display = 'none';
		},
		resetWardFilter(inputDisabled) {
			_selectWard = null;

			filterWardsInput.disabled = inputDisabled;
			filterWardsInput.value = "";
			filterWardsUl.innerHTML = '';
			filterWardsUl.style.display = 'none';
		},

		// Municipalities
		setMunicipalities(municipalities) {
			_currDistricts = municipalities;

			filterMunicipalitiesInput.disabled = false;
			filterMunicipalitiesInput.value = "";
			filterMunicipalitiesUl.innerHTML = '';
			filterMunicipalitiesUl.style.display = 'none';
		},
		fitlterMunicipalities(searchVal) {
			if (_currDistricts.length === 0) {
				return [];
			}

			searchVal = searchVal.trim();
			if (searchVal === "") {
				return _currDistricts;
			}

			return _currDistricts.filter(municipality => municipality.n.includes(searchVal));
		},

		// Districts
		setDistricts (districts) {
			_currDistricts = districts;
		},
		fitlterDistricts (searchVal) {
			if (_currDistricts.length === 0) {
				return [];
			}

			searchVal = searchVal.trim();
			if (searchVal === "") {
				return _currDistricts;
			}

			return _currDistricts.filter(district => district.n.includes(searchVal));
		},

		// Wards
		setWards(wards) {
			_currWards = wards;
		},
		fitlterWards(searchVal) {
			if (_currWards.length === 0) {
				return [];
			}

			searchVal = searchVal.trim();
			if (searchVal === "") {
				return _currWards;
			}

			return _currWards.filter(ward => ward.n.includes(searchVal));
		},

		populateFilter(inputElem, ulElem, opts, callback) {
			ulElem.innerHTML = '';

			opts.forEach(opt => {
				const li = document.createElement('li');
				li.value = opt.id;
				li.textContent = opt.n;
				li.addEventListener('click', () => {
					inputElem.value = opt.n;
					ulElem.style.display = 'none';
					callback(opt);
				});
				li.addEventListener('mouseover', () => {
					li.style.backgroundColor = '#f0f0f0';
				});
				li.addEventListener('mouseout', () => {
					li.style.backgroundColor = '';
				});
				ulElem.appendChild(li);
			});

			ulElem.style.display = opts.length ? 'block' : 'none';
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
			filterDistrictsWrapper: document.getElementById("districts-filter-wrapper"),
			filterDistrictsInput: document.getElementById("filter-districts-input"),
			filterDistrictsUl: document.getElementById("filter-districts-ul"),
			filterWardsWrapper: document.getElementById("wards-filter-wrapper"),
			filterWardsInput: document.getElementById("filter-wards-input"),
			filterWardsUl: document.getElementById("filter-wards-ul")
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
