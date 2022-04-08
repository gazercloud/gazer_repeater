<!-- Animation -->
function onEntry(entry) {
  entry.forEach(change => {
    if (change.isIntersecting) {
      change.target.classList.add('ani-show');
    }
  });
}

function onEntry2(entry) {
  entry.forEach(change => {
    if (change.isIntersecting) {
      change.target.classList.add('ani2-show');
    }
  });
}

function onEntry3(entry) {
  entry.forEach(change => {
    if (change.isIntersecting) {
      change.target.classList.add('ani3-show');
    }
  });
}

function onEntry4(entry) {
  entry.forEach(change => {
    if (change.isIntersecting) {
      change.target.classList.add('ani4-show');
    }
  });
}

function onEntry5(entry) {
  entry.forEach(change => {
    if (change.isIntersecting) {
      change.target.classList.add('ani5-show');
    }
  });
}

let options = { threshold: [0.25] };

let observer = new IntersectionObserver(onEntry, options);
let observer2 = new IntersectionObserver(onEntry2, options);
let observer3 = new IntersectionObserver(onEntry3, options);
let observer4 = new IntersectionObserver(onEntry4, options);
let observer5 = new IntersectionObserver(onEntry5, options);

let elements = document.querySelectorAll('.ani');
for (let elm of elements) {
  observer.observe(elm);
}

let elements2 = document.querySelectorAll('.ani2');
for (let elm of elements2) {
  observer2.observe(elm);
}

let elements3 = document.querySelectorAll('.ani3');
for (let elm of elements3) {
  observer3.observe(elm);
}

let elements4 = document.querySelectorAll('.ani4');
for (let elm of elements4) {
  observer4.observe(elm);
}

let elements5 = document.querySelectorAll('.ani5');
for (let elm of elements5) {
  observer5.observe(elm);
}
