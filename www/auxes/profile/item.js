'use strict';

const e = React.createElement;

class Item extends React.Component {
    constructor(props) {
        super(props);
        this.state = { liked: false, name: props.ItemName };
    }

    render() {
        return e(
            'button',
            { onClick: () => {} },
            'Item'
        );
    }
}
