'use strict';

const e = React.createElement;

function Comp(props) {
    return e(
        'button',
        { onClick: () => console.log("OK") },
        'Нравится'
    );
}

