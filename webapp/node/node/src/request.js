let activeRepeater = '';
let activeRepeaterProcessing = false;

export default async function Request(func, data = {}) {
    if (activeRepeater === '') {
        updateActiveHost()
        throw Error("waiting for the route to the node ...");
    }

    const formData = new FormData();
    formData.append('fn', func);
    formData.append('rj', JSON.stringify(data));

    // Default options are marked with *
    return await fetchWithTimeout('https://' + activeRepeater + '/api/request', {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'include', // include, *same-origin, omit
        headers: {
            //'Content-Type': 'multipart/form-data; boundary=CUSTOM'
            // 'Content-Type': 'application/x-www-form-urlencoded',
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *client
        body: formData // body data type must match "Content-Type" header
    }); // parses JSON response into native JavaScript objects
}

export function RequestFailed() {
    activeRepeater = ''
    updateActiveHost()
}

function updateActiveHost() {
    if (activeRepeater !== '') {
        return
    }

    if (activeRepeaterProcessing) {
        return
    }

    activeRepeaterProcessing = true
    requestActiveHost().then((res) => {
        if (res.status === 200) {
            res.json().then(
                (result) => {
                    activeRepeater = result.host
                }
            );
        } else {
            console.log("requestActiveHost error", res)
        }
        activeRepeaterProcessing = false
    }).catch(res => {
        activeRepeaterProcessing = false
    });
}

async function fetchWithTimeout(resource, options) {
    const { timeout = 5000 } = options;

    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);

    const response = await fetch(resource, {
        ...options,
        signal: controller.signal
    });
    clearTimeout(id);

    return response;
}

export async function requestActiveHost() {
    const formData = new FormData();

    let host = window.location.hostname
    let indexOfDot = host.indexOf("-n.gazer.cloud")
    let nodeId = host.substr(0, indexOfDot)
    console.log("requestActiveHost nodeId", nodeId)

    let data = {
        node_id: nodeId
    }

    formData.append('fn', "s-where-node");
    formData.append('rj', JSON.stringify(data));

    // Default options are marked with *
    return await fetchWithTimeout('https://home.gazer.cloud/api/request', {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'include', // include, *same-origin, omit
        headers: {
            //'Content-Type': 'multipart/form-data; boundary=CUSTOM'
            // 'Content-Type': 'application/x-www-form-urlencoded',
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *client
        body: formData // body data type must match "Content-Type" header
    }); // parses JSON response into native JavaScript objects
}

export async function RequestHome(func, data = {}) {
    const formData = new FormData();
    formData.append('fn', func);
    formData.append('rj', JSON.stringify(data));

    // Default options are marked with *
    return await fetchWithTimeout('/api/request', {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            //'Content-Type': 'multipart/form-data; boundary=CUSTOM'
            // 'Content-Type': 'application/x-www-form-urlencoded',
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *client
        body: formData // body data type must match "Content-Type" header
    }); // parses JSON response into native JavaScript objects
}
