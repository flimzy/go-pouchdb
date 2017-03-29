if ( $global.PouchDB === undefined ) {
    try {
        $global.PouchDB = require('pouchdb');
    } catch(e) {}
}
