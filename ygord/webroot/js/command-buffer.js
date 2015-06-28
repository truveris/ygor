/*
 * CommandBuffer wraps a list of commands to be executed by the channel
 * controller.
 */
function CommandBuffer() {
    this.buffer = [];
}

CommandBuffer.prototype.push = function(item) {
    if (item == "skip") {
        this.buffer = [];
        return;
    }
    this.buffer.push(item);
}

CommandBuffer.prototype.shift = function() {
    return this.buffer.shift()
}

CommandBuffer.prototype.shift = function() {
    return this.buffer.shift()
}
