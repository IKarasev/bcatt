let simulation;

// Настройки графа сети
const widthGraphBox = 1100;
const heightGraphBox = 400;

let graphNodes;
let graphLinks;

// Настройки графа цепочки блоков
const forkGrSettings = {
  width: 1100,
  height: 600,
  margin: { top: 40, right: 20, bottom: 20, left: 20 },
  blockSize: 60,
  blockPadding: 10,
  timelineHeight: 30,
  timelineToChainGap: 40,
  color: {
    block: "#7aa4ff",
    evil: "#ed5b51",
    blockStroke: "#000",
    evilStroke: "#000",
    heighlight: "#ffbb00",
    link: "#666",
    timeline: "#888",
    timelineStroke: "#aaa",
    text: "#000",
    evilttext: "#000",
  },
  font: {
    blockLabel: 14,
    timelineLabel: 12,
  },
};
forkGrSettings.layerSpacing =
  forkGrSettings.blockSize + forkGrSettings.blockPadding + 10;

// Создание контейнера графа сети
const mainGraphSvg = d3
  .select("#graph")
  .append("svg")
  .attr("width", "100%")
  .attr("height", "100%")
  .attr("viewBox", [0, 0, widthGraphBox, heightGraphBox])
  .style("cursor", "grab")
  .call(
    d3
      .zoom()
      .scaleExtent([0.1, 5])
      .on("start", () => svgGraph.style("cursor", "grabbing"))
      .on("end", () => svgGraph.style("cursor", "grab"))
      .on("zoom", zoomedGraph),
  );

const svgGraph = mainGraphSvg.append("g");

// Создание контейнеров цепочки
const forkGrSvg = d3
  .select("#chainfork")
  .append("svg")
  .attr("width", forkGrSettings.width)
  .attr("height", forkGrSettings.height)
  .attr("view-box", `0, 0, {boxWidth}`);

const forkGrZoomGroup = forkGrSvg.append("g");
forkGrSvg.call(
  d3
    .zoom()
    .scaleExtent([0.1, 10])
    .on("zoom", (event) => forkGrZoomGroup.attr("transform", event.transform)),
);

forkGrZoomGroup.append("g").attr("class", "timeline");
forkGrZoomGroup.append("g").attr("class", "links");
forkGrZoomGroup.append("g").attr("class", "blocks");

//**** Инициализация визуализаций
getCreateGraph();
getCreateForkGr();

//**** ВИЗУАЛИЗАЦИЯ СЕТИ
// Построение графа сети
function getCreateGraph() {
  fetch("/visual/net")
    .then((response) => {
      if (!response.ok) throw new Error("Network error");
      return response.json();
    })
    .then((data) => {
      createNetGraph(data);
    })
    .catch((error) => {
      console.error("Fetch error:", error);
    });
}

// Создание графа тополгии сети
async function createNetGraph(data) {
  // Задача цветовой шкалы
  const latencyColor = d3
    .scaleLinear()
    .domain([0, 100, 200])
    .range(["#2ca02c", "#ffd700", "#d62728"]);

  // Создание симуляции графа
  simulation = d3
    .forceSimulation(data.nodes)
    .force(
      "link",
      d3
        .forceLink(data.links)
        .id((d) => d.id)
        .distance(10),
    )
    .force("charge", d3.forceManyBody().strength(-500))
    .force("center", d3.forceCenter(widthGraphBox / 2, heightGraphBox / 2))
    .force("collision", d3.forceCollide().radius(50));

  // Создание ребер
  const link = svgGraph
    .append("g")
    .selectAll("line")
    .data(data.links)
    .join("line")
    .attr("class", "link")
    .attr("stroke", (d) => latencyColor(d.latency))
    .attr("stroke-width", (d) => Math.sqrt(d.latency) / 5);

  // Веса ребер
  const linkText = svgGraph
    .append("g")
    .selectAll("text")
    .data(data.links)
    .join("text")
    .attr("font-size", 10)
    .attr("fill", "#555")
    .text((d) => `${d.latency}ms`);

  // Группы узлов
  const node = svgGraph
    .append("g")
    .selectAll("g")
    .data(data.nodes)
    .join("g")
    .attr("class", (d) => `node ${d.group}`)
    .call(drag(simulation));

  // Обводка узлов
  node
    .append("circle")
    .attr("r", (d) => getNodeRadius(d.group))
    .attr("fill", (d) => getNodeColor(d.group))
    // .attr("class", (d) => `node ${d.group}`)
    // .attr("stroke", "#fff");
    .attr("stroke", (d) => (d.attacked ? "#a00" : "#fff"));

  // Подписи узлов
  node
    .append("text")
    .attr("dy", -15)
    .attr("text-anchor", "middle")
    .text((d) => d.id)
    .attr("fill", (d) => getNodeColor(d.group));

  // Обновление позиций
  simulation.on("tick", () => {
    link
      .attr("x1", (d) => d.source.x)
      .attr("y1", (d) => d.source.y)
      .attr("x2", (d) => d.target.x)
      .attr("y2", (d) => d.target.y);

    linkText
      .attr("x", (d) => (d.source.x + d.target.x) / 2)
      .attr("y", (d) => (d.source.y + d.target.y) / 2);

    node.attr("transform", (d) => `translate(${d.x},${d.y})`);
  });

  // Перемещение узлов
  function drag(simulation) {
    function dragstarted(event) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      event.subject.fx = event.subject.x;
      event.subject.fy = event.subject.y;
    }

    function dragged(event) {
      event.subject.fx = event.x;
      event.subject.fy = event.y;
    }

    function dragended(event) {
      if (!event.active) simulation.alphaTarget(0);
      event.subject.fx = null;
      event.subject.fy = null;
    }

    return d3
      .drag()
      .on("start", dragstarted)
      .on("drag", dragged)
      .on("end", dragended);
  }

  // Выделение подписи ребра при наведении
  link
    .on("mouseover", function (event, d) {
      d3.select(this).attr("stroke-width", Math.sqrt(d.latency) / 3);
      linkText.filter((dd) => dd === d).attr("font-weight", "bold");
    })
    .on("mouseout", function (event, d) {
      d3.select(this).attr("stroke-width", Math.sqrt(d.latency) / 5);
      linkText.filter((dd) => dd === d).attr("font-weight", "normal");
    });
}

// Обработчик событий изменения графа
function updateGraph(data) {
  if (!simulation) {
    getCreateGraph();
  } else {
    const updData = JSON.parse(data.data);
    if (updData.action == "miner") {
      updateGraphMiner(updData.nodes);
    } else if ((updData.action = "attacked")) {
      updateGraphAttacked(updData.nodes);
    }
  }
}

function updateGraphMiner(nodesData) {
  let nodes = svgGraph.selectAll(".node.miner");
  if (!nodes.empty()) {
    nodes.node().classList.replace("miner", "regular");
    nodes
      .select("circle")
      .attr("fill", getNodeColor("regular"))
      .attr("r", getNodeRadius("regular"));
  }

  nodes = svgGraph.selectAll(".node.regular");
  for (const n of nodesData) {
    node = nodes.filter((d) => d.id === n.id);
    node.node().classList.replace("regular", "miner");
    node
      .select("circle")
      .attr("fill", getNodeColor("miner"))
      .attr("r", getNodeRadius("miner"));
  }
  // simulation.alpha(0.3).restart();
}

function updateGraphAttacked(nodesData) {
  let nodes = svgGraph.selectAll(".node");
  for (const n of nodesData) {
    nodes
      .filter((d) => d.id === n.id)
      .node()
      .classList.add("attacked");
  }
}

// ***** Построение графа цепочки блоков

// Состояния цепочки блоков
const forkGrState = {
  blocks: new Map(), // hash -> block
  forks: new Map(), // hash -> y position (layer)
  heightForks: new Map(), // height -> list[hashes]
  forkLayers: 0,
};

// Создание грфафа цепочки
async function getCreateForkGr() {
  fetch("/visual/fork")
    .then((response) => {
      if (!response.ok) throw new Error("Fork fetch: network error");
      return response.json();
    })
    .then((data) => {
      data.forEach(forkAddBlock);
      forkRender();
    })
    .catch((error) => {
      console.error("Fetch error:", error);
    });
}

// Обновление цепочки блоков
async function forkUpdate(data) {
  const newBlock = JSON.parse(data.data);
  if (!forkGrState.blocks.has(newBlock.hash)) {
    forkAddBlock(newBlock);
    forkRender();
  }
}

// Добавление блока в цепочку
function forkAddBlock(block) {
  forkGrState.blocks.set(block.hash, block);

  // Определение слоя блока в цепочки
  const prev = forkGrState.blocks.get(block.prev);
  if (block.prev == "" || block.prev === "genesis") {
    forkGrState.forks.set(block.hash, 0);
    forkGrState.heightForks.set(block.height, [block.hash]);
  } else {
    let layer = forkGrState.forks.get(block.prev);
    const sameHeightForks = forkGrState.heightForks.get(block.height) || [];
    if (sameHeightForks.some((h) => forkGrState.forks.get(h) === layer)) {
      layer = ++forkGrState.forkLayers;
    }
    forkGrState.forks.set(block.hash, layer);
    forkGrState.heightForks.set(block.height, [...sameHeightForks, block.hash]);
  }
}

// Рендеринг цепочки блоков
function forkRender() {
  forkGrZoomGroup.select(".links").selectAll("*").remove();
  forkGrZoomGroup.select(".blocks").selectAll("*").remove();
  const maxTick = d3.max([...forkGrState.blocks.values()], (d) => d.tick);
  const xScale = d3
    .scaleLinear()
    .domain([0, maxTick + 1])
    .range([
      forkGrSettings.margin.left,
      (maxTick + 1) *
        (forkGrSettings.blockSize + forkGrSettings.blockPadding + 10),
    ]);

  // Таймлайн
  const ticks = [
    ...new Set([...forkGrState.blocks.values()].map((d) => d.tick)),
  ];
  forkRenderTimeline(xScale, ticks);

  // Связи
  const linkData = [...forkGrState.blocks.values()].filter(
    (b) => b.prev && forkGrState.blocks.has(b.prev),
  );

  const links = forkGrZoomGroup
    .select(".links")
    .selectAll("path")
    .data(linkData, (d) => d.hash);

  links
    .enter()
    .append("path")
    .attr("fill", "none")
    .attr("stroke", forkGrSettings.color.link)
    .attr("stroke-width", 1.5)
    .attr("d", (d) => {
      const parent = forkGrState.blocks.get(d.prev);
      const x1 = xScale(parent.tick) + forkGrSettings.blockSize / 2;
      const y1 = forkGetY(parent);
      const x2 = xScale(d.tick) + forkGrSettings.blockSize / 2;
      const y2 = forkGetY(d);
      const curveX = x1 == x2 ? x1 - (y2 - y1) / 2 : (x1 + x2) / 2;
      return `M${x1},${y1} C${curveX},${y1} ${curveX},${y2} ${x2},${y2}`;
    });

  // Блоки
  const blockGroups = forkGrZoomGroup
    .select(".blocks")
    .selectAll("g.block")
    .data([...forkGrState.blocks.values()], (d) => d.hash);

  const entered = blockGroups
    .enter()
    .append("g")
    .attr("class", "block")
    .attr(
      "transform",
      (d) =>
        `translate(${xScale(d.tick)}, ${forkGetY(d) - forkGrSettings.blockSize / 2})`,
    );

  entered
    .append("rect")
    .attr("rx", 10)
    .attr("ry", 10)
    .attr("width", forkGrSettings.blockSize)
    .attr("height", forkGrSettings.blockSize)
    .attr("fill", (d) =>
      d.evil ? forkGrSettings.color.evil : forkGrSettings.color.block,
    )
    .attr("stroke", forkGrSettings.color.blockStroke);

  entered
    .append("text")
    .attr("x", forkGrSettings.blockSize / 2)
    .attr("y", 20)
    .attr("text-anchor", "middle")
    .attr("fill", forkGrSettings.color.text)
    .attr("font-size", forkGrSettings.font.blockLabel + "px")
    .text((d) => `${d.height}`);

  entered
    .append("text")
    .attr("x", forkGrSettings.blockSize / 2)
    .attr("y", 40)
    .attr("text-anchor", "middle")
    .attr("fill", forkGrSettings.color.text)
    .attr("font-size", forkGrSettings.font.blockLabel - 2 + "px")
    .text((d) => d.node);

  entered
    .on("mouseover", function (event, d) {
      forkHighlightChain(d.hash);
    })
    .on("mouseout", function () {
      forkClearHighlight();
    });
}

// Рендеринг таймлайна
function forkRenderTimeline(xScale, ticks) {
  const timelineGroup = forkGrZoomGroup.select(".timeline");
  timelineGroup.selectAll("*").remove();

  // Ось
  timelineGroup
    .append("line")
    .attr("x1", forkGrSettings.margin.left)
    .attr(
      "x2",
      ticks.length *
        (forkGrSettings.blockSize + forkGrSettings.blockPadding + 10),
    )
    .attr("y1", forkGrSettings.margin.top + forkGrSettings.timelineHeight / 2)
    .attr("y2", forkGrSettings.margin.top + forkGrSettings.timelineHeight / 2)
    .attr("stroke", forkGrSettings.color.timeline)
    .attr("stroke-width", 2);

  // Тики и подписи
  timelineGroup
    .selectAll("line.tick")
    .data(ticks)
    .enter()
    .append("line")
    .attr("class", "tick")
    .attr("x1", (d) => xScale(d) + forkGrSettings.blockSize / 2)
    .attr("x2", (d) => xScale(d) + forkGrSettings.blockSize / 2)
    .attr(
      "y1",
      forkGrSettings.margin.top + forkGrSettings.timelineHeight / 2 - 5,
    )
    .attr(
      "y2",
      forkGrSettings.margin.top +
        forkGrSettings.timelineHeight +
        (forkGrState.forkLayers + 1) * forkGrSettings.layerSpacing,
    )
    .attr("stroke", forkGrSettings.color.timelineStroke)
    .attr("stroke-dasharray", "10,5")
    .attr("stroke-opacity", "0.5");

  timelineGroup
    .selectAll("text.tick-label")
    .data(ticks)
    .enter()
    .append("text")
    .attr("class", "tick-label")
    .attr("x", (d) => xScale(d) + forkGrSettings.blockSize / 2)
    .attr(
      "y",
      forkGrSettings.margin.top + forkGrSettings.timelineHeight / 2 - 8,
    )
    .attr("text-anchor", "middle")
    .attr("font-size", forkGrSettings.font.timelineLabel + "px")
    .text((d) => `Tick ${d}`);
}

// Вычисление y-координаты блока
function forkGetY(block) {
  let layer = forkGrState.forks.get(block.hash) || 0;
  layer = layer == 0 ? 0 : forkGrState.forkLayers - layer + 1;
  return (
    forkGrSettings.margin.top +
    forkGrSettings.timelineHeight +
    forkGrSettings.timelineToChainGap +
    layer * forkGrSettings.layerSpacing
  );
}

// Выделение цепочки блока при наведении
function forkHighlightChain(hash) {
  let currentHash = hash;
  while (currentHash && forkGrState.blocks.has(currentHash)) {
    forkGrZoomGroup
      .selectAll("g.block")
      .filter((b) => b.hash === currentHash)
      .select("rect")
      .attr("stroke", forkGrSettings.color.heighlight)
      .attr("stroke-width", 3);

    const block = forkGrState.blocks.get(currentHash);
    const parentHash = block.prev;

    if (parentHash && forkGrState.blocks.has(parentHash)) {
      forkGrZoomGroup
        .selectAll(".links path")
        .filter((d) => d.hash === currentHash)
        .attr("stroke", forkGrSettings.color.heighlight)
        .attr("stroke-width", 3);
    }

    currentHash = parentHash;
  }
}

// Очистка выделения при снятии фокуса
function forkClearHighlight() {
  forkGrZoomGroup
    .selectAll("g.block rect")
    .attr("stroke", forkGrSettings.color.blockStroke)
    .attr("stroke-width", 1);

  forkGrZoomGroup
    .selectAll(".links path")
    .attr("stroke", forkGrSettings.color.link)
    .attr("stroke-width", 1.5);
}

// ***** Вспомогательные функции
// Цвет ноды графа
function getNodeColor(group) {
  return {
    miner: "#ff9900",
    user: "#3366cc",
    regular: "#33cc33",
  }[group];
}

// Размер ноды
function getNodeRadius(group) {
  return {
    miner: 10,
    regular: 8,
    user: 6,
  }[group];
}

// Активация зумирования окна с графом
function zoomedGraph(event) {
  svgGraph.attr("transform", event.transform);
}

// Центрирование окна обозрения
function viewBoxCenter() {
  const bounds = svgGraph.node().getBBox();

  const contentCenterX = bounds.x + bounds.width / 2;
  const contentCenterY = bounds.y + bounds.height / 2;

  const viewCenterX = widthGraphBox / 2;
  const viewCenterY = heightGraphBox / 2;

  const translateX = viewCenterX - contentCenterX;
  const translateY = viewCenterY - contentCenterY;

  mainGraphSvg.call(
    forkZoom.transform,
    d3.zoomIdentity.translate(translateX, translateY).scale(1),
  );
}
