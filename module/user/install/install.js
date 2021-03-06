db.createCollection('Users');
db.Users.ensureIndex({
    name: 1
}, {
    unique: true
});

db.createCollection('WechatUserBindings');
db.WechatUserBindings.ensureIndex({
    openId: 1
}, {
    unique: true
});

db.Users.insert({
    name: 'admin',
    passwordDigest: '21232f297a57a5a743894a0e4a801fc3',
    nickname: '主办方管理员',
    mobile: '',
    flatPermissions: [
        'MERCHANT.STAFF.RETRIEVE',
        'MERCHANT.STAFF.MODIFY',
        'MERCHANT.RETRIEVE',
        'MERCHANT.MODIFY',
        'COMPETITION.DRAWN_TICKET.INSPECT',
        'COMPETITION.RETRIEVE',
        'COMPETITION.MODIFY',
        'USER.RETRIEVE',
        'USER.MODIFY',
        'USER.SPONSOR_MANAGER'
    ]
});
