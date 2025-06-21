const filteredCandidateContainer = document.getElementById(
  "filtered-candidate-container"
);
const shareContainer = document.getElementById("share-container");
const pepTalk = document.querySelector(`.pep-talk`);

const NewSearchCandidateHandler = (config) => {
  const layers = {
    municipality: {
      title: "縣市",
      inputView: config.filterMunicipalitiesInputView,
      inputTyping: config.filterMunicipalitiesInputTyping,
      ul: config.filterMunicipalitiesUl,
      wrapper: config.filterMunicipalitiesWrapper,
      backBtn: config.filterMunicipalitiesBackBtn,
      data: [],
      selected: null,
    },
    district: {
      title: "行政區",
      inputView: config.filterDistrictsInputView,
      inputTyping: config.filterDistrictsInputTyping,
      ul: config.filterDistrictsUl,
      wrapper: config.filterDistrictsWrapper,
      backBtn: config.filterDistrictsBackBtn,
      data: [],
      selected: null,
    },
    ward: {
      title: "鄉鎮村里",
      inputView: config.filterWardsInputView,
      inputTyping: config.filterWardsInputTyping,
      ul: config.filterWardsUl,
      wrapper: config.filterWardsWrapper,
      backBtn: config.filterWardsBackBtn,
      data: [],
      selected: null,
    },
  };

  function bindFilterEvents(level, clickFunc, goBackFunc) {
    const { inputView, inputTyping, ul, wrapper, backBtn } = layers[level];

    inputView.addEventListener("click", (e) => {
      const isMobile = window.matchMedia("(max-width: 600px)").matches;
      if (!isMobile) {
        initUlAndInputForDesktop();
      }

      inputView.style.display = "none";
      inputTyping.style.display = "block";
      inputTyping.value = inputView.value;

      populateFilter(
        inputView,
        inputTyping,
        ul,
        filterOptions(level, ""),
        clickFunc,
        wrapper
      );

      if (isMobile) setTimeout(() => wrapper.classList.add("open"), 10);
      if (!isMobile) {
        inputTyping.focus();
      }
    });
    inputTyping.addEventListener("input", (e) => {
      const isMobile = window.matchMedia("(max-width: 600px)").matches;
      populateFilter(
        inputView,
        inputTyping,
        ul,
        filterOptions(level, e.target.value.trim()),
        clickFunc,
        wrapper
      );

      if (isMobile) setTimeout(() => wrapper.classList.add("open"), 10);
    });
    inputTyping.addEventListener("keydown", (e) => {
      const items = ul.querySelectorAll("li");
      if (!items.length) return;

      let index = Array.from(items).findIndex((li) =>
        li.classList.contains("highlighted")
      );

      if (e.key === "ArrowDown") {
        e.preventDefault();
        if (index === -1) index = 0;
        else index = Math.min(index + 1, items.length - 1);

        items.forEach((item) => item.classList.remove("highlighted"));
        items[index].classList.add("highlighted");

        ensureVisible(ul, items[index]);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        if (index === -1) index = 0;
        else index = Math.max(index - 1, 0);

        items.forEach((item) => item.classList.remove("highlighted"));
        items[index].classList.add("highlighted");

        ensureVisible(ul, items[index]);
      } else if (e.key === "Enter" && index >= 0) {
        e.preventDefault();
        items[index].click();
      }
    });

    backBtn.addEventListener("click", () => {
      goBackFunc(true);
      wrapper.classList.remove("open");
    });

    document.addEventListener("click", (e) => {
      const otherWrappers = Object.values(layers)
        .filter((layer) => layer.wrapper !== wrapper)
        .map((layer) => layer.wrapper);
      if (
        !wrapper.contains(e.target) &&
        !otherWrappers.some((w) => w.contains(e.target))
      ) {
        goBackFunc(false);
      }
    });
  }

  function initUlAndInputForDesktop() {
    Object.keys(layers).forEach((key) => {
      layers[key].ul.innerHTML = "";
      layers[key].ul.style.display = "none";
      layers[key].inputTyping.value = "";
      layers[key].inputTyping.style.display = "none";
      layers[key].inputView.style.display = "block";
    });
  }

  function ensureVisible(ul, element) {
    const ulRect = ul.getBoundingClientRect();
    const elRect = element.getBoundingClientRect();

    if (elRect.bottom > ulRect.bottom) {
      ul.scrollTop += elRect.bottom - ulRect.bottom;
    } else if (elRect.top < ulRect.top) {
      ul.scrollTop += elRect.top - ulRect.top;
    }
  }

  function generateSynonymVariants(str) {
    const variants = new Set([str]);
    const synonymMap = {
      台: ["臺"],
      臺: ["台"],
    };

    for (const [key, values] of Object.entries(synonymMap)) {
      if (str.includes(key)) {
        values.forEach((val) => {
          variants.add(str.replace(new RegExp(key, "g"), val));
        });
      }
    }

    return Array.from(variants);
  }

  function filterOptions(level, searchTerm) {
    const opts = layers[level].data;

    if (opts.length === 0) return [];
    if (searchTerm === "") return opts;

    const searchTermVariants = generateSynonymVariants(
      searchTerm.toLowerCase()
    );

    filteredOpts = opts.filter((opt) => {
      const optNameVariants = generateSynonymVariants(opt.n.toLowerCase());

      return searchTermVariants.some((searchVariant) =>
        optNameVariants.some((optVariant) => optVariant.includes(searchVariant))
      );
    });

    if (!filteredOpts.length) {
      return [
        {
          id: null,
          n: `請輸入正確${layers[level].title}`,
        },
      ];
    }

    return filteredOpts;
  }

  function populateFilter(
    inputView,
    inputTyping,
    ul,
    opts,
    onClick,
    wrapperElem
  ) {
    ul.innerHTML = "";

    opts.forEach((opt, idx) => {
      const li = document.createElement("li");
      li.value = opt.id;
      li.textContent = opt.n;

      if (opt.id !== null) {
        li.addEventListener("click", () => {
          inputView.value = opt.n;
          ul.style.display = "none";
          onClick(opt);

          inputView.style.display = "block";
          inputTyping.style.display = "none";
          wrapperElem.classList.remove("open");
        });

        li.addEventListener("pointerdown", (e) => {
          ul.querySelectorAll("li").forEach((item) =>
            item.classList.remove("hover")
          );
          li.classList.add("hover");
        });
        li.addEventListener("pointerup", () => {
          li.classList.remove("hover");
        });
        li.addEventListener("pointercancel", () => {
          li.classList.remove("hover");
        });

        li.addEventListener("mouseover", () => {
          li.classList.add("hover");
        });
        li.addEventListener("mouseout", () => {
          li.classList.remove("hover");
        });
      } else {
        li.style.color = "#999";
        li.style.cursor = "not-allowed";
      }

      ul.appendChild(li);
    });

    ul.style.display = opts.length ? "block" : "none";
  }

  return {
    init() {
      mask.classList.add("active");
      shareContainer.style.display = "none";
      sendListMunicipalitiesReq()
        .then((data) => {
          if (!Object.hasOwn(data, "result")) {
            showShareContainer();
          } else if (Object.hasOwn(data.result, "divisions")) {
            this.setMunicipalities(data.result.divisions);
            this.resetMunicipalityFilter();
          } else {
            console.error("invalid district");
          }
        })
        .catch((error) => console.error(error))
        .finally(() => mask.classList.remove("active"));

      bindFilterEvents(
        "municipality",
        this.selectMunicipality.bind(this),
        this.goBackMunicipality.bind(this)
      );
      bindFilterEvents(
        "district",
        this.selectDistrict.bind(this),
        this.goBackDistrict.bind(this)
      );
      bindFilterEvents(
        "ward",
        this.selectWard.bind(this),
        this.goBackWard.bind(this)
      );
    },

    // Municipalities, Districts, Wards Selectors
    selectMunicipality(municipality) {
      layers.municipality.selected = municipality;

      this.setDistricts([]);
      this.resetDistrictFilter(false);

      mask.classList.add("active");
      filteredCandidateContainer.innerHTML = "";
      shareContainer.style.display = "none";
      sendSearchConstituenciesReq(layers.municipality.selected.id, null, null)
        .then((data) => {
          if (!Object.hasOwn(data, "result")) {
            showShareContainer();
          } else if (Object.hasOwn(data.result, "divisions")) {
            this.setDistricts(data.result.divisions);
            this.resetDistrictFilter(false);
          } else {
            console.error("invalid municipality");
          }
        })
        .catch((error) => console.error(error))
        .finally(() => mask.classList.remove("active"));
    },
    selectDistrict(district) {
      layers.district.selected = district;

      this.setWards([]);
      this.resetWardFilter(false);

      mask.classList.add("active");
      filteredCandidateContainer.innerHTML = "";
      shareContainer.style.display = "none";
      sendSearchConstituenciesReq(
        layers.municipality.selected.id,
        layers.district.selected.id,
        null
      )
        .then((data) => {
          if (!Object.hasOwn(data, "result")) {
            showShareContainer();
          } else if (Object.hasOwn(data.result, "divisions")) {
            this.setWards(data.result.divisions);
            this.resetWardFilter(false);
          } else {
            console.error("invalid district");
          }
        })
        .catch((error) => console.error(error))
        .finally(() => mask.classList.remove("active"));
    },
    selectWard(ward) {
      layers.ward.selected = ward;

      mask.classList.add("active");
      filteredCandidateContainer.innerHTML = "";
      shareContainer.style.display = "none";
      sendSearchConstituenciesReq(
        layers.municipality.selected.id,
        layers.district.selected.id,
        layers.ward.selected.id
      )
        .then((data) => {
          if (!Object.hasOwn(data, "result")) {
            showShareContainer();
          } else if (Object.hasOwn(data.result, "legislators")) {
            const address = `${layers.municipality.selected.n}${layers.district.selected.n}${layers.ward.selected.n}`;
            showFilteredCandidateContainer(data.result.legislators, address);
          } else {
            console.error("invalid ward");
          }
        })
        .catch((error) => console.error(error))
        .finally(() => mask.classList.remove("active"));
    },

    // Municipalities, Districts, Wards Reset
    resetMunicipalityFilter() {
      layers.municipality.inputView.disabled = false;
      layers.municipality.inputView.value = "";
      layers.municipality.inputTyping.disabled = false;
      layers.municipality.inputTyping.value = "";
      layers.municipality.ul.innerHTML = "";
      layers.municipality.ul.style.display = "none";

      this.resetDistrictFilter(true);
      this.resetWardFilter(true);
    },
    resetDistrictFilter(inputDisabled) {
      layers.district.selected = null;

      layers.district.inputView.disabled = !layers.district.data.length
        ? true
        : inputDisabled;
      layers.district.inputView.value = "";
      layers.district.inputTyping.disabled = !layers.district.data.length
        ? true
        : inputDisabled;
      layers.district.inputTyping.value = "";
      layers.district.ul.innerHTML = "";
      layers.district.ul.style.display = "none";

      this.resetWardFilter(true);
    },
    resetWardFilter(inputDisabled) {
      layers.ward.selected = null;

      layers.ward.inputView.disabled = !layers.ward.data.length
        ? true
        : inputDisabled;
      layers.ward.inputView.value = "";
      layers.ward.inputTyping.disabled = !layers.ward.data.length
        ? true
        : inputDisabled;
      layers.ward.inputTyping.value = "";
      layers.ward.ul.innerHTML = "";
      layers.ward.ul.style.display = "none";
    },

    // Municipalities, Districts, Wards GoBack
    goBackMunicipality(forGoBack) {
      layers.municipality.inputView.style.display = "block";
      layers.municipality.inputTyping.style.display = "none";
      if (forGoBack) {
        layers.municipality.inputView.disabled = false;
        layers.municipality.inputTyping.disabled = false;
      }
      layers.municipality.inputView.value = layers.municipality.selected
        ? layers.municipality.selected.n
        : "";
      layers.municipality.ul.innerHTML = "";
      layers.municipality.ul.style.display = "none";
    },
    goBackDistrict(forGoBack) {
      layers.district.inputView.style.display = "block";
      layers.district.inputTyping.style.display = "none";
      if (forGoBack) {
        layers.district.inputView.disabled = layers.district.data.length
          ? false
          : true;
        layers.district.inputTyping.disabled = layers.district.data.length
          ? false
          : true;
      }
      layers.district.inputView.value = layers.district.selected
        ? layers.district.selected.n
        : "";
      layers.district.ul.innerHTML = "";
      layers.district.ul.style.display = "none";
    },
    goBackWard(forGoBack) {
      layers.ward.inputView.style.display = "block";
      layers.ward.inputTyping.style.display = "none";
      if (forGoBack) {
        layers.ward.inputView.disabled = layers.ward.data.length ? false : true;
        layers.ward.inputTyping.disabled = layers.ward.data.length
          ? false
          : true;
      }
      layers.ward.inputView.value = layers.ward.selected
        ? layers.ward.selected.n
        : "";
      layers.ward.ul.innerHTML = "";
      layers.ward.ul.style.display = "none";
    },

    // Municipalities, Districts, Wards Setters
    setMunicipalities(municipalities) {
      layers.municipality.data = municipalities;
    },
    setDistricts(districts) {
      layers.district.data = districts;
    },
    setWards(wards) {
      layers.ward.data = wards;
    },
  };
};

async function sendListMunicipalitiesReq() {
  const fullUrl = `${baseURL}/apis/municipalities`;

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

  const fullUrl = `${baseURL}/apis/constituencies?${params.toString()}`;

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
  const initMunicipalityHasFailed =
    document.querySelector(
      `.municipalities ul[data-city="1"] li.recall-failed`
    ) !== null
      ? true
      : false;
  if (initMunicipalityHasFailed) {
    pepTalk.style.display = "flex";
  }

  const searchHdr = NewSearchCandidateHandler({
    filterMunicipalitiesWrapper: document.getElementById(
      "municipalities-filter-wrapper"
    ),
    filterMunicipalitiesInputView: document.getElementById(
      "filter-municipalities-input-view"
    ),
    filterMunicipalitiesInputTyping: document.getElementById(
      "filter-municipalities-input-typing"
    ),
    filterMunicipalitiesUl: document.getElementById("filter-municipalities-ul"),
    filterMunicipalitiesBackBtn: document.getElementById(
      "filter-municipalities-back-btn"
    ),

    filterDistrictsWrapper: document.getElementById("districts-filter-wrapper"),
    filterDistrictsInputView: document.getElementById(
      "filter-districts-input-view"
    ),
    filterDistrictsInputTyping: document.getElementById(
      "filter-districts-input-typing"
    ),
    filterDistrictsUl: document.getElementById("filter-districts-ul"),
    filterDistrictsBackBtn: document.getElementById(
      "filter-districts-back-btn"
    ),

    filterWardsWrapper: document.getElementById("wards-filter-wrapper"),
    filterWardsInputView: document.getElementById("filter-wards-input-view"),
    filterWardsInputTyping: document.getElementById(
      "filter-wards-input-typing"
    ),
    filterWardsUl: document.getElementById("filter-wards-ul"),
    filterWardsBackBtn: document.getElementById("filter-wards-back-btn"),
  });
  searchHdr.init();

  dialogMask.addEventListener("click", function (event) {
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
    legislators.forEach((legislator) => {
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
          recallStages = [1, 2, 3]
            .map(
              (stage) => `
						<h4 class="recall-stage ${stage === legislator.recallStage ? "active" : ""}">
							<span>第 ${stage} 階段</span>${stage === 3 ? "罷免投票" : "連署罷免"}
						</h4>
						${stage < 3 ? '<span class="icon-step-arrow"></span>' : ""}
					`
            )
            .join("");

          recallStages = `<div class="recall-stage-flow">${recallStages}</div>`;

          if (legislator.recallStage === 1 || legislator.recallStage === 2) {
            if (legislator.formDeployed) {
              candidateAction = `<a href="${legislator.participateURL}?address=${address}"><button class="btn-primary lg w100">連署罷免</button></a>`;
            } else {
              candidateAction = `<button class="btn-primary lg w100" disabled>${legislator.recallStage} 階準備中</button>`;
            }
          } else {
            candidateAction = `
			<a href="${legislator.participateURL}"><button class="btn-primary lg w100">回家投票！馬上訂車票</button></a>
			<a href="${legislator.calendarURL}" target="_blank"><button class="btn-black lg w100">加入 Google 日曆提醒投票</button></a>
			`;
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
				</div>
				<div class="candidate-action">
					<div class="urgency">
						<div class="days-left">
							${
                legislator.daysLeft >= 0
                  ? `<i class="icon-urgent"></i><i class="icon-urgent"></i><i class="icon-urgent"></i>造冊中，請儘速繳交！`
                  : legislator.recallStatus === "ONGOING"
                  ? legislator.isShortage
                    ? `<i class="icon-urgent"></i><i class="icon-urgent"></i>
					<i class="icon-urgent"></i><i class="icon-urgent"></i><i class="icon-urgent"></i>儘速補件！！`
                    : legislator.votingDate
                    ? `<span style="font-size: 1.2em;">${legislator.votingDateStr}<br>請回戶籍地投票！</span>`
                    : `尚待中選會公告`
                  : "罷免連署未達成門檻。別灰心，我們還是需要您的力量，支持其他選區進行中的罷免活動，幫忙分享資訊！"
              }
						</div>
					</div>
					${candidateAction}
				</div>
				${
          legislator.recallStatus === "ONGOING"
            ? "<p>現行的公職人員罷免制度為分2階段徵求選民連署，通過之後第3階段再進行選民投票決定罷免結果。請大家務必3階段全部跟到最後一刻！</p>"
            : ""
        }`;
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

new Swiper(".swiper", {
  slidesPerView: 1.05,
  spaceBetween: 16,
  pagination: { el: ".swiper-pagination", clickable: true },
  autoplay: {
    delay: 3000,
    disableOnInteraction: false,
  },
  loop: true,
});

const municipalityLists = document.querySelectorAll(".municipalities ul");
const municipalityTags = document.querySelectorAll(".municipality-tag");

function toggleCityList(cityId) {
  const targetUl = document.querySelector(`ul[data-city="${cityId}"]`);
  const targetTag = document.querySelector(
    `.municipality-tag[data-city="${cityId}"]`
  );
  const hasFailed =
    targetUl.querySelector(`li.recall-failed`) !== null ? true : false;

  if (targetUl && targetUl.style.display === "flex") {
    return;
  }

  municipalityLists.forEach((ul) => (ul.style.display = "none"));

  municipalityTags.forEach((tag) => {
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
