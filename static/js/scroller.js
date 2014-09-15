var Scroller = {
    init: function() {
        var hidWidth;
        var scrollBarWidths = 40;

        var widthOfList = function() {
            var itemsWidth = 0;
            $('.tabs-list li').each(function() {
                var itemWidth = $(this).outerWidth();
                itemsWidth += itemWidth;
            });
            return itemsWidth;
        };

        var widthOfHidden = function() {
            return (($('.tabs-wrapper').outerWidth()) - widthOfList() - getLeftPosi()) - scrollBarWidths;
        };

        var getLeftPosi = function() {
            return $('.tabs-list').position().left;
        };

        var reAdjust = function() {
            if (($('.tabs-wrapper').outerWidth()) < widthOfList()) {
                $('.scroller-right').show();
            } else {
                $('.scroller-right').hide();
            }

            if (getLeftPosi() < 0) {
                $('.scroller-left').show();
            } else {
                $('.item').animate({
                    left: "-=" + getLeftPosi() + "px"
                }, 'slow');
                $('.scroller-left').hide();
            }
        }

        reAdjust();

        $(window).on('resize', function(e) {
            reAdjust();
        });

        $('.scroller-right').click(function() {

            $('.scroller-left').fadeIn('slow');
            $('.scroller-right').fadeOut('slow');

            $('.tabs-list').animate({
                left: "+=" + widthOfHidden() + "px"
            }, 'slow', function() {

            });
        });

        $('.scroller-left').click(function() {

            $('.scroller-right').fadeIn('slow');
            $('.scroller-left').fadeOut('slow');

            $('.tabs-list').animate({
                left: "-=" + getLeftPosi() + "px"
            }, 'slow', function() {

            });
        });
    }
}
