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
